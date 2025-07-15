package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
)

// screenshotPattern is the regex pattern used to match macOS screenshot filenames
const screenshotPattern = "^(Screen Shot|Screenshot) \\d{4}-\\d{2}-\\d{2} at \\d{1,2}\\.\\d{2}\\.\\d{2}[\\s\u202F]*(AM|PM)\\.png$"

// processScreenshot processes a single file, moving it if it matches the screenshot pattern
func processScreenshot(filename, srcDir, destDir string, dryRun bool, re *regexp.Regexp) {
	// Check if the file name matches the screenshot pattern
	if !re.MatchString(filename) {
		return // Not a screenshot, do nothing
	}

	srcPath := filepath.Join(srcDir, filename)
	destPath := filepath.Join(destDir, filename)

	if dryRun {
		fmt.Printf("[DRY RUN] Would move %s to %s\n", filename, destDir)
	} else {
		// Check if source file exists before trying to move it
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			log.Printf("Source file does not exist: %s", srcPath)
			return
		}

		// Move (rename) the file to the destination folder
		err := os.Rename(srcPath, destPath)
		if err != nil {
			log.Printf("Failed to move file %s: %v (srcPath: %s, destPath: %s)", filename, err, srcPath, destPath)
		} else {
			fmt.Printf("Moved %s to %s\n", filename, destDir)
		}
	}
}

// processExistingFiles processes all files currently in the source directory
func processExistingFiles(srcDir, destDir string, dryRun bool, re *regexp.Regexp) {
	// Read all items in the source folder
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		log.Printf("Failed to read source folder: %v", err)
		return
	}

	// Loop over each entry in the source folder
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			fmt.Printf("Skipping directory %s\n", entry.Name())
			continue
		}

		// Process the file
		processScreenshot(entry.Name(), srcDir, destDir, dryRun, re)
	}
}

// watchFolder watches the source directory for new files and processes screenshots
func watchFolder(ctx context.Context, srcDir, destDir string, dryRun bool, re *regexp.Regexp, processed chan bool) {
	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Add the source directory to the watcher
	err = watcher.Add(srcDir)
	if err != nil {
		log.Printf("Failed to add directory to watcher: %v", err)
		return
	}

	fmt.Printf("Watching %s for new screenshots...\n", srcDir)

	// Event deduplication: track recently processed files
	recentFiles := make(map[string]time.Time)
	const dedupeWindow = 2 * time.Second
	const cleanupInterval = 30 * time.Second
	lastCleanup := time.Now()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop watching
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			
			// Only process Create events (new files)
			if event.Op&fsnotify.Create == fsnotify.Create {
				filename := filepath.Base(event.Name)
				
				// Periodic cleanup of old entries
				if time.Since(lastCleanup) > cleanupInterval {
					cutoff := time.Now().Add(-dedupeWindow)
					for file, timestamp := range recentFiles {
						if timestamp.Before(cutoff) {
							delete(recentFiles, file)
						}
					}
					lastCleanup = time.Now()
				}
				
				// Check for duplicate events
				if lastProcessed, exists := recentFiles[filename]; exists {
					if time.Since(lastProcessed) < dedupeWindow {
						continue // Skip duplicate
					}
				}
				recentFiles[filename] = time.Now()
				
				// Add a small delay to ensure file is fully written
				time.Sleep(100 * time.Millisecond)
				
				// Process the new file
				processScreenshot(filename, srcDir, destDir, dryRun, re)
				
				// Signal that processing is complete (for tests)
				if processed != nil {
					select {
					case processed <- true:
					default:
						// Channel is full, don't block
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func main() {
	// Define command-line flags for source and destination folders.
	srcFolder := flag.String("src", "", "Source folder to search for screenshots")
	destFolder := flag.String("dest", "", "Destination folder to move screenshots")
	dryRun := flag.Bool("dry-run", false, "Show what would be done without actually moving files")
	watch := flag.Bool("watch", false, "Watch the source folder for new screenshots and move them automatically")

	flag.Parse()
	// Validate that both source and destination folders are provided.
	if *srcFolder == "" || *destFolder == "" {
		fmt.Println("Usage: go run main.go -src=<source folder> -dest=<destination folder> [-dry-run] [-watch]")
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

	// Always process existing files first
	processExistingFiles(*srcFolder, *destFolder, *dryRun, re)

	// If watch mode is enabled, start watching for new files
	if *watch {
		ctx := context.Background()
		watchFolder(ctx, *srcFolder, *destFolder, *dryRun, re, nil)
	}
}
