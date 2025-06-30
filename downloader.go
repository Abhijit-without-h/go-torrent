package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type ChunkRequest struct {
	Index int
	Peer  string
}

type ChunkResult struct {
	Index int
	Data  []byte
	Error error
}

type DownloadStats struct {
	mu              sync.RWMutex
	TotalChunks     int
	CompletedChunks int
	FailedChunks    int
	BytesDownloaded int64
	StartTime       time.Time
}

func (d *DownloadStats) IncrementCompleted(bytes int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.CompletedChunks++
	d.BytesDownloaded += int64(bytes)
}

func (d *DownloadStats) IncrementFailed() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.FailedChunks++
}

func (d *DownloadStats) GetProgress() (int, int, int, int64, float64) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	progress := float64(d.CompletedChunks) / float64(d.TotalChunks) * 100
	return d.TotalChunks, d.CompletedChunks, d.FailedChunks, d.BytesDownloaded, progress
}

func StartDownloader(torrentPath string) {
	file, err := os.Open(torrentPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var torrent Torrent
	err = json.NewDecoder(file).Decode(&torrent)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting download: %s (%d chunks, %.2f MB)\n", 
		torrent.FileName, torrent.TotalChunks, float64(torrent.FileSize)/(1024*1024))

	os.MkdirAll("./downloads", os.ModePerm)

	outputPath := "./downloads/" + torrent.FileName
	output, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	err = output.Truncate(torrent.FileSize)
	if err != nil {
		panic(err)
	}

	stats := &DownloadStats{
		TotalChunks: torrent.TotalChunks,
		StartTime:   time.Now(),
	}

	go reportDownloadProgress(stats)

	completedChunks := make([]bool, torrent.TotalChunks)
	var completedMu sync.RWMutex
	
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("Retry attempt %d/%d\n", attempt+1, maxRetries)
		}
		
		if downloadAllChunks(output, &torrent, stats, completedChunks, &completedMu) {
			break
		}
		
		if attempt == maxRetries-1 {
			fmt.Println("Max retries reached, download may be incomplete")
		}
	}

	output.Sync()

	completedMu.RLock()
	missingChunks := 0
	for i, completed := range completedChunks {
		if !completed {
			fmt.Printf("Missing chunk: %d\n", i)
			missingChunks++
		}
	}
	completedMu.RUnlock()

	total, completed, failed, bytes, progress := stats.GetProgress()
	duration := time.Since(stats.StartTime)
	
	fmt.Printf("\nDownload completed!\n")
	fmt.Printf("Chunks: %d/%d completed, %d failed, %d missing\n", 
		completed, total, failed, missingChunks)
	fmt.Printf("Size: %.2f MB\n", float64(bytes)/(1024*1024))
	fmt.Printf("Time: %v\n", duration.Round(time.Second))
	if duration.Seconds() > 0 {
		fmt.Printf("Average speed: %.2f MB/s\n", float64(bytes)/(1024*1024)/duration.Seconds())
	}
	fmt.Printf("Progress: %.1f%%\n", progress)
	fmt.Printf("File saved to: %s\n", outputPath)
	
	if missingChunks > 0 {
		fmt.Printf("WARNING: %d chunks are missing, file may be corrupted\n", missingChunks)
	}
}

func downloadAllChunks(output *os.File, torrent *Torrent, stats *DownloadStats, 
	completedChunks []bool, completedMu *sync.RWMutex) bool {
	
	const maxConcurrency = 5
	chunkRequests := make(chan ChunkRequest, 100)
	chunkResults := make(chan ChunkResult, 100)

	var wg sync.WaitGroup
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go downloadWorker(&wg, chunkRequests, chunkResults, torrent)
	}

	go func() {
		defer close(chunkRequests)
		
		completedMu.RLock()
		for i := 0; i < torrent.TotalChunks; i++ {
			if !completedChunks[i] {
				for _, peer := range torrent.Peers {
					select {
					case chunkRequests <- ChunkRequest{Index: i, Peer: peer}:
					default:
					}
				}
			}
		}
		completedMu.RUnlock()
	}()

	go func() {
		wg.Wait()
		close(chunkResults)
	}()

	processedResults := make(map[int]bool)
	
	for result := range chunkResults {
		if processedResults[result.Index] {
			continue
		}

		completedMu.RLock()
		if completedChunks[result.Index] {
			completedMu.RUnlock()
			continue
		}
		completedMu.RUnlock()

		if result.Error != nil {
			fmt.Printf("Failed to download chunk %d: %v\n", result.Index, result.Error)
			stats.IncrementFailed()
			continue
		}

		hash := sha256.Sum256(result.Data)
		expectedHash := torrent.Hashes[result.Index]
		actualHash := hex.EncodeToString(hash[:])
		
		if actualHash != expectedHash {
			fmt.Printf("Hash mismatch for chunk %d: expected %s, got %s\n", 
				result.Index, expectedHash[:16], actualHash[:16])
			stats.IncrementFailed()
			continue
		}

		offset := int64(result.Index * torrent.ChunkSize)
		n, err := output.WriteAt(result.Data, offset)
		if err != nil {
			fmt.Printf("Failed to write chunk %d: %v\n", result.Index, err)
			stats.IncrementFailed()
			continue
		}

		if n != len(result.Data) {
			fmt.Printf("Incomplete write for chunk %d: wrote %d of %d bytes\n", 
				result.Index, n, len(result.Data))
			stats.IncrementFailed()
			continue
		}

		completedMu.Lock()
		completedChunks[result.Index] = true
		completedMu.Unlock()
		
		processedResults[result.Index] = true
		stats.IncrementCompleted(len(result.Data))
		
		fmt.Printf("Downloaded chunk %d (%d bytes)\n", result.Index, len(result.Data))
	}

	completedMu.RLock()
	allComplete := true
	for _, completed := range completedChunks {
		if !completed {
			allComplete = false
			break
		}
	}
	completedMu.RUnlock()

	return allComplete
}

func downloadWorker(wg *sync.WaitGroup, requests <-chan ChunkRequest, results chan<- ChunkResult, torrent *Torrent) {
	defer wg.Done()

	for req := range requests {
		data, err := downloadChunk(req.Peer, req.Index, torrent.ChunkSize)
		
		select {
		case results <- ChunkResult{
			Index: req.Index,
			Data:  data,
			Error: err,
		}:
		case <-time.After(1 * time.Second):
			fmt.Printf("Timeout sending result for chunk %d\n", req.Index)
		}
	}
}

func downloadChunk(peer string, chunkIndex int, chunkSize int) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", peer, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	_, err = fmt.Fprintf(conn, "%d\n", chunkIndex)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	buf := make([]byte, chunkSize)
	totalRead := 0
	
	for totalRead < chunkSize {
		n, err := conn.Read(buf[totalRead:])
		if err != nil {
			if totalRead > 0 {
				break
			}
			return nil, fmt.Errorf("read failed: %v", err)
		}
		totalRead += n
		if n == 0 {
			break
		}
	}

	if totalRead == 0 {
		return nil, fmt.Errorf("no data received")
	}

	return buf[:totalRead], nil
}

func reportDownloadProgress(stats *DownloadStats) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		total, completed, failed, bytes, progress := stats.GetProgress()
		if completed == total {
			return
		}
		
		duration := time.Since(stats.StartTime)
		var speed float64
		if duration.Seconds() > 0 {
			speed = float64(bytes) / (1024 * 1024) / duration.Seconds()
		}
		
		fmt.Printf("Progress: %.1f%% (%d/%d chunks, %d failed, %.2f MB/s)\n",
			progress, completed, total, failed, speed)
	}
}