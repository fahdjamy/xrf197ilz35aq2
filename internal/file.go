package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func CloseFileWithRetry(file io.Closer, maxRetries int, retryAfter time.Duration) error {
	var err error
	for retries := 0; retries < maxRetries; retries++ {
		err = file.Close()
		if err == nil {
			return nil // Success!
		}

		fmt.Println("Error closing file (attempt", retries+1, "):", err)
		time.Sleep(retryAfter)
	}

	// If we reach here, all retries failed
	return fmt.Errorf("failed to close file after %d retries: %v", maxRetries, err)
}

func OpenFromRoot(relativePath string) (*os.File, error) {
	_, filename, _, ok := runtime.Caller(0) // Get caller file info. 0 = current function
	if !ok {
		return nil, fmt.Errorf("failed to get caller information")
	}

	// Get the directory of the *current* file (the file containing this function).
	thisDir := filepath.Dir(filename)

	// Find the project root by going up from thisDir as many times as needed.
	//  This part is project-structure-dependent.  Here are a few examples:

	// I.E: the internal package `internal` is directly under the root:
	rootDir := filepath.Dir(thisDir) // Go one level up.

	// Construct the absolute path to the target file.
	absolutePath := filepath.Join(rootDir, relativePath)

	// Open the file.
	file, err := os.Open(absolutePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func IsDir(relativePath string) (bool, error) {
	_, filename, _, ok := runtime.Caller(0) // Get caller file info. 0 = current function
	if !ok {
		return false, fmt.Errorf("failed to get caller information")
	}
	thisDir := filepath.Dir(filename)
	rootDir := filepath.Dir(thisDir)
	absolutePath := filepath.Join(rootDir, relativePath)

	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}
