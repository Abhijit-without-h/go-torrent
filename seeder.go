package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SeederStats struct {
	mu           sync.RWMutex
	ChunksServed int
	BytesServed  int64
	Connections  int
	StartTime    time.Time
}

func (s *SeederStats) IncrementChunks(bytes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ChunksServed++
	s.BytesServed += int64(bytes)
}

func (s *SeederStats) AddConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Connections++
}

func (s *SeederStats) RemoveConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Connections--
}

func (s *SeederStats) GetStats() (int, int64, int, time.Duration) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ChunksServed, s.BytesServed, s.Connections, time.Since(s.StartTime)
}

var seederStats = &SeederStats{StartTime: time.Now()}

func StartSeeder(filepath string) {
	torrent := CreateTorrentFile(filepath, 64*1024, []string{"127.0.0.1:8080"})
	fmt.Printf("Seeding: %s (%d chunks)\n", torrent.FileName, torrent.TotalChunks)

	go reportSeederStats()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Println("Seeder listening on :8080")

	timeout := time.After(1 * time.Minute) // Stop after 1 minutes
    for {
        select {
        case <-timeout:
            fmt.Println("Seeder timeout reached, shutting down.")
            return
        default:
            conn, err := ln.Accept()
            if err != nil {
                fmt.Println("Connection error:", err)
                continue
            }
            seederStats.AddConnection()
            go handlePeerConnection(conn, filepath, torrent)
        }
	}
}

func handlePeerConnection(conn net.Conn, filepath string, torrent Torrent) {
	defer func() {
		conn.Close()
		seederStats.RemoveConnection()
	}()

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)
	chunkIndexStr, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	chunkIndex, err := strconv.Atoi(strings.TrimSpace(chunkIndexStr))
	if err != nil {
		return
	}

	if chunkIndex < 0 || chunkIndex >= torrent.TotalChunks {
		return
	}

	f, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer f.Close()

	offset := int64(chunkIndex * torrent.ChunkSize)
	_, err = f.Seek(offset, 0)
	if err != nil {
		return
	}

	buf := make([]byte, torrent.ChunkSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return
	}

	_, err = conn.Write(buf[:n])
	if err != nil {
		return
	}

	seederStats.IncrementChunks(n)
}

func reportSeederStats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		chunks, bytes, connections, uptime := seederStats.GetStats()
		fmt.Printf("Stats: %d chunks served, %.2f MB uploaded, %d active connections, uptime: %v\n",
			chunks, float64(bytes)/(1024*1024), connections, uptime.Round(time.Second))
	}
}