package throttle_test

import (
	"context"
	"testing"
	"time"

	"github.com/tonymet/gcloud-go/throttle"
)

const ttl time.Duration = 10 * time.Millisecond

func runThrottleTest(t *testing.T, cap int, testFunc func(th throttle.Throttle, testChan chan int)) {
	ctx, cancel := context.WithTimeout(t.Context(), ttl)
	defer cancel()
	total := 0
	th := throttle.NewThrottle(cap)
	// count the number of test routines with plenty of buffer
	testChan := make(chan int, cap+1)
	testFunc(th, testChan)
	for {
		select {
		case val := <-testChan:
			total += val
		case <-ctx.Done():
			if total > cap {
				t.Fatalf("processed total %d exceeded cap %d", total, cap)
			}
			t.Log("we stayed within cap")
			return
		}
	}
}

func TestThrottle_Wait(t *testing.T) {
	cap := 5
	runThrottleTest(t, cap, func(th throttle.Throttle, testChan chan int) {
		// launch excessive go routines
		for range cap + 5 {
			go func() {
				th.Wait()
				defer th.Done()
				testChan <- 1
				time.Sleep(3 * ttl)
			}()
		}
	})
}

func TestThrottle_Go(t *testing.T) {
	cap := 5
	runThrottleTest(t, cap, func(th throttle.Throttle, testChan chan int) {
		for range cap + 5 {
			th.Go(func() {
				testChan <- 1
				time.Sleep(3 * ttl)
			})
		}
	})
}

func TestThrottle_Done(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), ttl)
	defer cancel()
	total := 0
	cap := 5
	th := throttle.NewThrottle(cap)
	// count the number of test routines with plenty of buffer
	testChan := make(chan int, cap+1)
	// launch excessive go routines
	for range cap {
		go func() {
			th.Wait()
			defer th.Done()
			testChan <- 1
			time.Sleep(3 * ttl)
		}()
	}
	// free one more
	th.Done()
	for {
		select {
		case val := <-testChan:
			total += val
		case <-ctx.Done():
			if total < cap {
				t.Fatalf("cap %d exceeded total %d ", cap, total)
			}
			t.Log("we processed all")
			return
		}
	}
}
