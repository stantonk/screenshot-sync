package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	// Define command-line flags for source and destination folders.
	srcFolder := flag.String("src", "", "Source folder to search for screenshots")
	destFolder := flag.String("dest", "", "Destination folder to move screenshots")

	// Validate that both source and destination folders are provided.
	if *srcFolder == "" || *destFolder == "" {
		fmt.Println("Usage: go run main.go -src=<source folder> -dest=<destination folder>")
		return
	}

	// Check that the source folder exists.
	if stat, err := os.Stat(*srcFolder); os.IsNotExist(err) || !stat.IsDir() {
		log.Fatalf("Source folder does not exist or is not a directory: %s", *srcFolder)
	}

	// Create the destination folder if it does not exist.
	if stat, err := os.Stat(*destFolder); os.IsNotExist(err) || !stat.IsDir() {
		err := os.MkdirAll(*destFolder, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create destination folder: %v", err)
		}
	}

	// Define a regular expression to match filenames of the format:
	// "Screenshot YYYY-MM-DD at H.MM.SS AM|PM.png"
	// For example: "Screenshot 2025-03-28 at 4.34.27 PM.png"
	pattern := `^Screenshot \d{4}-\d{2}-\d{2} at \d{1,2}\.\d{2}\.\d{2}\s*(AM|PM)\.png$`
	re := regexp.MustCompile(pattern)

	// Read all items in the source folder.
	entries, err := os.ReadDir(*srcFolder)
	if err != nil {
		log.Fatalf("Failed to read source folder: %v", err)
	}

	// Loop over each entry in the source folder.
	for _, entry := range entries {
		// Skip directories.
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		// Check if the file name matches the screenshot pattern.
		if re.MatchString(fileName) {
			srcPath := filepath.Join(*srcFolder, fileName)
			destPath := filepath.Join(*destFolder, fileName)

			// Move (rename) the file to the destination folder.
			err := os.Rename(srcPath, destPath)
			if err != nil {
				log.Printf("Failed to move file %s: %v", fileName, err)
			} else {
				fmt.Printf("Moved %s to %s\n", fileName, *destFolder)
			}
		}
	}
}
