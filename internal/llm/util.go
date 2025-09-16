package llm

import (
	"log"
	"strings"
	"sync"
)

// ProcessInChunks is a generic function that processes a slice of items in parallel chunks.
// It takes a slice of items, a chunk size, and a worker function to process each chunk.
// It returns a slice of results from the successful chunks and a map of failed items, keyed by the error reason.
func ProcessInChunks[T any, R any](
	items []T,
	chunkSize int,
	worker func(chunk []T) (R, error),
) ([]R, map[string][]T) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]R, 0)
	failedChunks := make(map[string][]T)

	for i := 0; i < len(items); i += chunkSize {
		end := min(i+chunkSize, len(items))
		chunk := items[i:end]
		wg.Add(1)

		go func(chunk []T) {
			defer wg.Done()

			result, err := worker(chunk)
			if err != nil {
				log.Printf("A chunk failed during parallel processing: %v", err)

				// Determine a simple error key for bucketing failed chunks.
				errorKey := "Processing Error"
				if strings.Contains(err.Error(), "DEADLINE_EXCEEDED") {
					errorKey = "Processing Timeout"
				} else if strings.Contains(strings.ToLower(err.Error()), "json") {
					errorKey = "Invalid Response"
				}

				mu.Lock()
				failedChunks[errorKey] = append(failedChunks[errorKey], chunk...)
				mu.Unlock()
				return
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(chunk)
	}

	wg.Wait()

	return results, failedChunks
}

func min(a, b int) int {
	if a < b {
			return a
		}
	return b
}