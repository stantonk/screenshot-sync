package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// screenshotPattern is the regex pattern used to match macOS screenshot filenames
const screenshotPattern = "^(Screen Shot|Screenshot) \\d{4}-\\d{2}-\\d{2} at \\d{1,2}\\.\\d{2}\\.\\d{2}[\\s\u202F]*(AM|PM)\\.png$"

func main() {
	// Define command-line flags for source and destination folders.
	srcFolder := flag.String("src", "", "Source folder to search for screenshots")
	destFolder := flag.String("dest", "", "Destination folder to move screenshots")
	dryRun := flag.Bool("dry-run", false, "Show what would be done without actually moving files")

	flag.Parse()
	// Validate that both source and destination folders are provided.
	if *srcFolder == "" || *destFolder == "" {
		fmt.Println("Usage: go run main.go -src=<source folder> -dest=<destination folder> [-dry-run]")
		return
	}

	if *dryRun {
		fmt.Println("Running in dry run mode")
	}

	// Check that the source folder exists.
	if stat, err := os.Stat(*srcFolder); os.IsNotExist(err) || !stat.IsDir() {
		log.Fatalf("Source folder does not exist or is not a directory: %s", *srcFolder)
	}

	// Create the destination folder if it does not exist.
	if !*dryRun {
		if stat, err := os.Stat(*destFolder); os.IsNotExist(err) || !stat.IsDir() {
			err := os.MkdirAll(*destFolder, os.ModePerm)
			if err != nil {
				log.Fatalf("Failed to create destination folder: %v", err)
			}
		}
	}

	re := regexp.MustCompile(screenshotPattern)

	// Read all items in the source folder.
	entries, err := os.ReadDir(*srcFolder)
	if err != nil {
		log.Fatalf("Failed to read source folder: %v", err)
	}

	// Loop over each entry in the source folder.
	for _, entry := range entries {
		// Skip directories.
		if entry.IsDir() {
			fmt.Printf("Skipping directory %s\n", entry.Name())
			continue
		}

		fileName := entry.Name()
		// Check if the file name matches the screenshot pattern.
		if re.MatchString(fileName) {
			srcPath := filepath.Join(*srcFolder, fileName)
			destPath := filepath.Join(*destFolder, fileName)

			if *dryRun {
				fmt.Printf("[DRY RUN] Would move %s to %s\n", fileName, *destFolder)
			} else {
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
}
