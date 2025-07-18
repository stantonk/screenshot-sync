# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a simple Go command-line tool that automatically organizes macOS screenshots by moving them from a source directory to a destination directory. The tool identifies screenshot files using regex pattern matching based on macOS naming conventions.

## Development Commands

- **Build**: `go build`
- **Run**: `go run main.go -src=<source> -dest=<destination> [-dry-run]`
- **Test**: `go test`
- **Run single test**: `go test -run TestScreenshotRegex`

## Architecture

The application consists of a single main package with:
- **main.go**: Core application logic with command-line flag parsing, file system operations, and regex matching
- **main_test.go**: Unit tests for the screenshot filename regex pattern
- **screenshotPattern**: Regex constant that matches both "Screen Shot" and "Screenshot" prefixes with macOS timestamp format, including support for narrow no-break spaces

The regex pattern handles edge cases like narrow no-break space characters (U+202F) that can appear in screenshot filenames.

## Key Implementation Details

- Uses `os.Rename()` for moving files (atomic operation on same filesystem)
- Creates destination directory automatically if it doesn't exist
- Supports dry-run mode for safe testing
- Skips directories in source folder
- Handles both regular spaces and narrow no-break spaces in filenames