package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run . <mode> <filepath>")
		fmt.Println("Modes: seed, download, create")
		os.Exit(1)
	}

	mode := os.Args[1]
	filepath := os.Args[2]

	switch mode {
	case "create":
		torrent := CreateTorrentFile(filepath, 64*1024, []string{"127.0.0.1:8080"})
		fmt.Printf("Created torrent: %s.torrent\n", torrent.FileName)
	case "seed":
		StartSeeder(filepath)
	case "download":
		StartDownloader(filepath)
	default:
		fmt.Println("Invalid mode. Use: create, seed, or download")
		os.Exit(1)
	}
}