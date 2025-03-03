package internal

import (
	"fmt"
	"io"
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
