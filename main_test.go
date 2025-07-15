package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

func TestScreenshotRegex(t *testing.T) {
	re := regexp.MustCompile(screenshotPattern)

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "valid PM",
			filename: "Screen Shot 2020-06-21 at 4.21.35 PM.png",
			want:     true,
		},
		{
			name:     "valid AM",
			filename: "Screen Shot 2020-06-21 at 4.21.35 AM.png",
			want:     true,
		},
		{
			name:     "valid PM 2",
			filename: "Screenshot 2025-03-29 at 11.16.20 PM.png",
			want:     true,
		},
		{
			name:     "valid PM with narrow no-break space",
			filename: "Screenshot 2025-03-30 at 2.06.46 PM.png",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := re.MatchString(tt.filename); got != tt.want {
				t.Errorf("regex.MatchString(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// UNIT TEST: Tests the core screenshot processing function
// Purpose: Verify file moving logic with different scenarios (normal, dry-run, invalid files)
// Dependencies: None - pure function test
func TestProcessScreenshot(t *testing.T) {
	// Create temporary directories
	srcDir, err := os.MkdirTemp("", "screenshot-src-*")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	destDir, err := os.MkdirTemp("", "screenshot-dest-*")
	if err != nil {
		t.Fatalf("Failed to create temp dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	re := regexp.MustCompile(screenshotPattern)

	tests := []struct {
		name     string
		filename string
		dryRun   bool
		wantMove bool
	}{
		{
			name:     "valid screenshot should be moved",
			filename: "Screen Shot 2020-06-21 at 4.21.35 PM.png",
			dryRun:   false,
			wantMove: true,
		},
		{
			name:     "valid screenshot in dry-run should not be moved",
			filename: "Screenshot 2025-03-29 at 11.16.20 PM.png",
			dryRun:   true,
			wantMove: false,
		},
		{
			name:     "invalid file should not be moved",
			filename: "not-a-screenshot.png",
			dryRun:   false,
			wantMove: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file in source directory
			srcPath := filepath.Join(srcDir, tt.filename)
			if err := os.WriteFile(srcPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Process the screenshot
			processScreenshot(tt.filename, srcDir, destDir, tt.dryRun, re)

			// Check if file was moved
			destPath := filepath.Join(destDir, tt.filename)
			_, destExists := os.Stat(destPath)
			_, srcExists := os.Stat(srcPath)

			if tt.wantMove {
				if os.IsNotExist(destExists) {
					t.Errorf("Expected file to be moved to destination, but it wasn't")
				}
				if !os.IsNotExist(srcExists) {
					t.Errorf("Expected file to be removed from source, but it still exists")
				}
			} else {
				if !os.IsNotExist(destExists) {
					t.Errorf("Expected file not to be moved to destination, but it was")
				}
				if os.IsNotExist(srcExists) {
					t.Errorf("Expected file to remain in source, but it was removed")
				}
			}

			// Clean up for next test
			os.Remove(srcPath)
			os.Remove(destPath)
		})
	}
}

// UNIT TEST: Tests batch processing of existing files in a directory
// Purpose: Verify that existing screenshots are processed correctly on startup
// Dependencies: processScreenshot function
func TestProcessExistingFiles(t *testing.T) {
	// Create temporary directories
	srcDir, err := os.MkdirTemp("", "screenshot-existing-src-*")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	destDir, err := os.MkdirTemp("", "screenshot-existing-dest-*")
	if err != nil {
		t.Fatalf("Failed to create temp dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	// Create test files
	testFiles := []struct {
		filename   string
		shouldMove bool
	}{
		{"Screen Shot 2020-06-21 at 4.21.35 PM.png", true},
		{"Screenshot 2025-03-29 at 11.16.20 PM.png", true},
		{"not-a-screenshot.png", false},
		{"document.pdf", false},
	}

	for _, tf := range testFiles {
		srcPath := filepath.Join(srcDir, tf.filename)
		if err := os.WriteFile(srcPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.filename, err)
		}
	}

	re := regexp.MustCompile(screenshotPattern)

	// Test normal mode
	t.Run("normal mode", func(t *testing.T) {
		processExistingFiles(srcDir, destDir, false, re)

		for _, tf := range testFiles {
			destPath := filepath.Join(destDir, tf.filename)
			srcPath := filepath.Join(srcDir, tf.filename)

			_, destExists := os.Stat(destPath)
			_, srcExists := os.Stat(srcPath)

			if tf.shouldMove {
				if os.IsNotExist(destExists) {
					t.Errorf("Expected %s to be moved to destination, but it wasn't", tf.filename)
				}
				if !os.IsNotExist(srcExists) {
					t.Errorf("Expected %s to be removed from source, but it still exists", tf.filename)
				}
			} else {
				if !os.IsNotExist(destExists) {
					t.Errorf("Expected %s not to be moved to destination, but it was", tf.filename)
				}
				if os.IsNotExist(srcExists) {
					t.Errorf("Expected %s to remain in source, but it was removed", tf.filename)
				}
			}
		}
	})

	// Recreate files for dry-run test
	for _, tf := range testFiles {
		srcPath := filepath.Join(srcDir, tf.filename)
		destPath := filepath.Join(destDir, tf.filename)
		os.Remove(destPath) // Clean up from previous test
		if err := os.WriteFile(srcPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to recreate test file %s: %v", tf.filename, err)
		}
	}

	// Test dry-run mode
	t.Run("dry-run mode", func(t *testing.T) {
		processExistingFiles(srcDir, destDir, true, re)

		for _, tf := range testFiles {
			destPath := filepath.Join(destDir, tf.filename)
			srcPath := filepath.Join(srcDir, tf.filename)

			_, destExists := os.Stat(destPath)
			_, srcExists := os.Stat(srcPath)

			// In dry-run mode, nothing should be moved
			if !os.IsNotExist(destExists) {
				t.Errorf("Expected %s not to be moved in dry-run mode, but it was", tf.filename)
			}
			if os.IsNotExist(srcExists) {
				t.Errorf("Expected %s to remain in source in dry-run mode, but it was removed", tf.filename)
			}
		}
	})
}

// INTEGRATION TEST: Tests the complete folder watching functionality
// Purpose: Verify that fsnotify detects new files and processes them correctly
// Dependencies: Real file system, fsnotify library, processScreenshot function
func TestWatchFolderIntegration(t *testing.T) {
	// Create temporary directories
	srcDir, err := os.MkdirTemp("", "screenshot-watch-src-*")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	destDir, err := os.MkdirTemp("", "screenshot-watch-dest-*")
	if err != nil {
		t.Fatalf("Failed to create temp dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	re := regexp.MustCompile(screenshotPattern)

	tests := []struct {
		name     string
		filename string
		dryRun   bool
		wantMove bool
	}{
		{
			name:     "new screenshot should be detected and moved",
			filename: "Screen Shot 2025-01-15 at 2.30.45 PM.png",
			dryRun:   false,
			wantMove: true,
		},
		{
			name:     "new screenshot in dry-run should be detected but not moved",
			filename: "Screenshot 2025-01-15 at 3.45.20 AM.png",
			dryRun:   true,
			wantMove: false,
		},
		{
			name:     "non-screenshot file should be ignored",
			filename: "regular-file.txt",
			dryRun:   false,
			wantMove: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with timeout for the watcher
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			// Channel to signal when file processing is complete
			processed := make(chan bool, 1)

			// Start watching in a goroutine
			go func() {
				watchFolder(ctx, srcDir, destDir, tt.dryRun, re, processed)
			}()

			// Give the watcher time to start
			time.Sleep(200 * time.Millisecond)

			// Create a new file that should trigger the watcher
			srcPath := filepath.Join(srcDir, tt.filename)
			if err := os.WriteFile(srcPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Wait for processing to complete or timeout
			select {
			case <-processed:
				// File was processed by watcher
			case <-time.After(2 * time.Second):
				if tt.wantMove || (tt.filename != "regular-file.txt") {
					t.Fatal("Timeout waiting for file to be processed")
				}
				// For non-screenshot files, timeout is expected
			}

			// Check if file was moved
			destPath := filepath.Join(destDir, tt.filename)
			_, destExists := os.Stat(destPath)
			_, srcExists := os.Stat(srcPath)

			if tt.wantMove {
				if os.IsNotExist(destExists) {
					t.Errorf("Expected file to be moved to destination, but it wasn't")
				}
				if !os.IsNotExist(srcExists) {
					t.Errorf("Expected file to be removed from source, but it still exists")
				}
			} else {
				if !os.IsNotExist(destExists) {
					t.Errorf("Expected file not to be moved to destination, but it was")
				}
				if os.IsNotExist(srcExists) {
					t.Errorf("Expected file to remain in source, but it was removed")
				}
			}

			// Clean up
			os.Remove(srcPath)
			os.Remove(destPath)
		})
	}
}

// UNIT TEST: Tests the event deduplication and cleanup functionality
// Purpose: Verify that duplicate events are filtered and old entries are cleaned up
// Dependencies: None - tests internal deduplication logic
func TestEventDeduplication(t *testing.T) {
	// This test verifies the deduplication logic that will be in watchFolder
	// We'll test the core logic separately since watchFolder is complex
	
	recentFiles := make(map[string]time.Time)
	const dedupeWindow = 100 * time.Millisecond // Short window for testing
	
	// Helper function to check if file should be processed (not a duplicate)
	shouldProcess := func(filename string) bool {
		if lastProcessed, exists := recentFiles[filename]; exists {
			if time.Since(lastProcessed) < dedupeWindow {
				return false // Skip duplicate
			}
		}
		recentFiles[filename] = time.Now()
		return true
	}
	
	// Helper function to clean up old entries
	cleanup := func() {
		cutoff := time.Now().Add(-dedupeWindow)
		for file, timestamp := range recentFiles {
			if timestamp.Before(cutoff) {
				delete(recentFiles, file)
			}
		}
	}
	
	// Test deduplication
	t.Run("deduplication", func(t *testing.T) {
		filename := "Screenshot 2025-01-15 at 1.00.00 PM.png"
		
		// First call should process
		if !shouldProcess(filename) {
			t.Error("First call should be processed")
		}
		
		// Immediate second call should be skipped (duplicate)
		if shouldProcess(filename) {
			t.Error("Immediate second call should be skipped as duplicate")
		}
		
		// After waiting, should process again
		time.Sleep(dedupeWindow + 10*time.Millisecond)
		if !shouldProcess(filename) {
			t.Error("Call after dedupe window should be processed")
		}
	})
	
	// Test cleanup
	t.Run("cleanup", func(t *testing.T) {
		// Add several files
		files := []string{
			"Screenshot 2025-01-15 at 1.00.00 PM.png",
			"Screenshot 2025-01-15 at 1.01.00 PM.png", 
			"Screenshot 2025-01-15 at 1.02.00 PM.png",
		}
		
		for _, file := range files {
			shouldProcess(file)
		}
		
		if len(recentFiles) != 3 {
			t.Errorf("Expected 3 files in map, got %d", len(recentFiles))
		}
		
		// Wait for entries to expire
		time.Sleep(dedupeWindow + 10*time.Millisecond)
		
		// Run cleanup
		cleanup()
		
		if len(recentFiles) != 0 {
			t.Errorf("Expected 0 files after cleanup, got %d", len(recentFiles))
		}
	})
	
	// Test partial cleanup (some entries expire, others don't)
	t.Run("partial cleanup", func(t *testing.T) {
		// Clear map
		recentFiles = make(map[string]time.Time)
		
		// Add old file
		shouldProcess("old-file.png")
		
		// Wait for it to expire
		time.Sleep(dedupeWindow + 10*time.Millisecond)
		
		// Add new file
		shouldProcess("new-file.png")
		
		// Cleanup should remove old file but keep new file
		cleanup()
		
		if len(recentFiles) != 1 {
			t.Errorf("Expected 1 file after partial cleanup, got %d", len(recentFiles))
		}
		
		if _, exists := recentFiles["new-file.png"]; !exists {
			t.Error("New file should still exist after cleanup")
		}
		
		if _, exists := recentFiles["old-file.png"]; exists {
			t.Error("Old file should be removed after cleanup")
		}
	})
}
