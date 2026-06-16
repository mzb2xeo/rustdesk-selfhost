package service

import (
	"sync"
	"testing"
)

// TestGetAppVersion
func TestGetAppVersion(t *testing.T) {
	s := &AppService{}
	v := s.GetAppVersion()
	// Print results
	t.Logf("App Version: %s", v)
}

func TestMultipleGetAppVersion(t *testing.T) {
	s := &AppService{}
	//Concurrency testing
	// Use WaitGroup to wait for all goroutines to complete
	wg := sync.WaitGroup{}
	wg.Add(10) // Start 10 goroutines
	// Start 10 goroutines
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done() // Decrement count after completion
			v := s.GetAppVersion()
			// Print results
			t.Logf("App Version: %s", v)
		}()
	}
	// Wait for all goroutines to complete
	wg.Wait()
}
