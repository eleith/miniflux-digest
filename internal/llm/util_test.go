package llm

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessInChunks(t *testing.T) {
	t.Run("happy path - processes all items", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5, 6, 7}
		chunkSize := 3

		worker := func(chunk []int) (string, error) {
			var sum int
			for _, item := range chunk {
				sum += item
			}
			return fmt.Sprintf("sum-%d", sum), nil
		}

		results, failedItems := ProcessInChunks(items, chunkSize, worker)

		assert.Empty(t, failedItems)
		assert.Len(t, results, 3) // 3 chunks: (1,2,3), (4,5,6), (7)
		assert.Contains(t, results, "sum-6")
		assert.Contains(t, results, "sum-15")
		assert.Contains(t, results, "sum-7")
	})

	t.Run("one chunk fails", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		chunkSize := 4

		worker := func(chunk []int) (string, error) {
			// Fail the second chunk
			if chunk[0] == 5 {
				return "", errors.New("processing error")
			}
			return "success", nil
		}

		results, failedChunks := ProcessInChunks(items, chunkSize, worker)

		assert.Len(t, results, 2) // Chunks 1 and 3 should succeed
		assert.Len(t, failedChunks, 1) // One failure reason
		assert.Contains(t, failedChunks, "Processing Error")
		assert.Len(t, failedChunks["Processing Error"], 4) // Chunk 2 (items 5,6,7,8) should fail
		assert.Equal(t, 5, failedChunks["Processing Error"][0])
	})

	t.Run("all chunks fail", func(t *testing.T) {
		items := []int{1, 2, 3, 4}
		chunkSize := 2

		worker := func(chunk []int) (string, error) {
			return "", errors.New("always fail")
		}

		results, failedChunks := ProcessInChunks(items, chunkSize, worker)

		assert.Empty(t, results)
		assert.Len(t, failedChunks, 1)
		assert.Len(t, failedChunks["Processing Error"], 4)
	})

	t.Run("items less than chunk size", func(t *testing.T) {
		items := []int{1, 2}
		chunkSize := 5

		worker := func(chunk []int) (string, error) {
			return "success", nil
		}

		results, failedItems := ProcessInChunks(items, chunkSize, worker)

		assert.Empty(t, failedItems)
		assert.Len(t, results, 1)
	})

	t.Run("zero items", func(t *testing.T) {
		items := []int{}
		chunkSize := 5

		worker := func(chunk []int) (string, error) {
			t.Fatal("worker should not be called for zero items")
			return "", nil
		}

		results, failedItems := ProcessInChunks(items, chunkSize, worker)

		assert.Empty(t, failedItems)
		assert.Empty(t, results)
	})

	t.Run("exact multiple of chunk size", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5, 6}
		chunkSize := 3

		worker := func(chunk []int) (string, error) {
			return "success", nil
		}

		results, failedItems := ProcessInChunks(items, chunkSize, worker)

		assert.Empty(t, failedItems)
		assert.Len(t, results, 2)
	})

	t.Run("thread safety", func(t *testing.T) {
		numItems := 1000
		items := make([]int, numItems)
		for i := 0; i < numItems; i++ {
			items[i] = i
		}
		chunkSize := 10

		var mu sync.Mutex
		var callCount int

		worker := func(chunk []int) (int, error) {
			mu.Lock()
			callCount++
			// Fail every 10th chunk
			if callCount%10 == 0 {
				mu.Unlock()
				return 0, errors.New("failed")
			}
			mu.Unlock()
			return len(chunk), nil
		}

		results, failedChunks := ProcessInChunks(items, chunkSize, worker)

		numChunks := numItems / chunkSize
		numFailedChunks := numChunks / 10
		numSuccessfulChunks := numChunks - numFailedChunks

		assert.Len(t, results, numSuccessfulChunks)
		assert.Len(t, failedChunks["Processing Error"], numFailedChunks*chunkSize)
	})
}