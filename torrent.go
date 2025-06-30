package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Torrent struct {
	FileName    string   `json:"file_name"`
	FileSize    int64    `json:"file_size"`
	ChunkSize   int      `json:"chunk_size"`
	Hashes      []string `json:"hashes"`
	Peers       []string `json:"peers"`
	CreatedAt   int64    `json:"created_at"`
	TotalChunks int      `json:"total_chunks"`
}

func CreateTorrentFile(path string, chunkSize int, peers []string) Torrent {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	totalChunks := int((fileInfo.Size() + int64(chunkSize) - 1) / int64(chunkSize))
	hashes := make([]string, 0, totalChunks)

	buf := make([]byte, chunkSize)
	hasher := sha256.New()

	for {
		n, err := file.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		hasher.Reset()
		hasher.Write(buf[:n])
		hash := hex.EncodeToString(hasher.Sum(nil))
		hashes = append(hashes, hash)
	}

	torrent := Torrent{
		FileName:    filepath.Base(path),
		FileSize:    fileInfo.Size(),
		ChunkSize:   chunkSize,
		Hashes:      hashes,
		Peers:       peers,
		CreatedAt:   time.Now().Unix(),
		TotalChunks: len(hashes),
	}

	os.MkdirAll("./torrents", os.ModePerm)

	jsonFile, err := os.Create("./torrents/" + torrent.FileName + ".torrent")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(torrent)
	if err != nil {
		panic(err)
	}

	return torrent
}