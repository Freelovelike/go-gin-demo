package services

import (
	"math"
	"sync"
	"testing"
)

func TestEventProbability(t *testing.T) {
	// Zero/negative inputs yield zero probability.
	for _, tc := range []struct{ rate, mult, cycles float64 }{
		{0, 1, 10}, {0.1, 0, 10}, {0.1, 1, 0}, {-1, 1, 10},
	} {
		if got := eventProbability(tc.rate, tc.mult, tc.cycles); got != 0 {
			t.Errorf("eventProbability(%v,%v,%v) = %v, want 0", tc.rate, tc.mult, tc.cycles, got)
		}
	}

	// Monotonic in cycles: more elapsed time => higher chance.
	p1 := eventProbability(0.1, 1, 10)
	p2 := eventProbability(0.1, 1, 100)
	if !(p1 > 0 && p1 < p2 && p2 < 1) {
		t.Errorf("expected 0 < p1(%v) < p2(%v) < 1", p1, p2)
	}

	// Matches the closed form 1-(1-p)^n it replaced (no 1000-cycle cap).
	pPerCycle := 0.1 * (10.0 / 3600.0) * 1.0
	want := 1 - math.Pow(1-pPerCycle, 5000)
	if got := eventProbability(0.1, 1, 5000); math.Abs(got-want) > 1e-9 {
		t.Errorf("eventProbability(0.1,1,5000) = %v, want %v", got, want)
	}

	// Probability is clamped to [0,1] even when per-cycle p>=1.
	if got := eventProbability(1000, 1, 1); got != 1 {
		t.Errorf("saturated rate = %v, want 1", got)
	}
}

func TestLockUserSerializes(t *testing.T) {
	// Concurrent increments under the same user lock must not race.
	const goroutines = 50
	const perG = 100
	counter := 0
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perG; j++ {
				unlock := lockUser(1)
				counter++ // read-modify-write, racy without the lock
				unlock()
			}
		}()
	}
	wg.Wait()
	if counter != goroutines*perG {
		t.Errorf("counter = %d, want %d", counter, goroutines*perG)
	}
}

func TestLockUserPerUserIndependent(t *testing.T) {
	// Different user IDs get independent locks (no false sharing / deadlock).
	u1 := lockUser(100)
	u2 := lockUser(200) // would block forever if they shared a mutex
	u2()
	u1()
}
