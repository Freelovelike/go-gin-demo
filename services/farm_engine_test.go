package services

import (
	"math"
	"sync"
	"testing"
)

func TestEventProbability(t *testing.T) {
	// 零/负输入产生零概率。
	for _, tc := range []struct{ rate, mult, cycles float64 }{
		{0, 1, 10}, {0.1, 0, 10}, {0.1, 1, 0}, {-1, 1, 10},
	} {
		if got := eventProbability(tc.rate, tc.mult, tc.cycles); got != 0 {
			t.Errorf("eventProbability(%v,%v,%v) = %v, want 0", tc.rate, tc.mult, tc.cycles, got)
		}
	}

	// 在周期中单调递增：经过的时间越长 => 概率越高。
	p1 := eventProbability(0.1, 1, 10)
	p2 := eventProbability(0.1, 1, 100)
	if !(p1 > 0 && p1 < p2 && p2 < 1) {
		t.Errorf("expected 0 < p1(%v) < p2(%v) < 1", p1, p2)
	}

	// 匹配它取代的闭式 1-(1-p)^n（没有 1000 个周期的上限）。
	pPerCycle := 0.1 * (10.0 / 3600.0) * 1.0
	want := 1 - math.Pow(1-pPerCycle, 5000)
	if got := eventProbability(0.1, 1, 5000); math.Abs(got-want) > 1e-9 {
		t.Errorf("eventProbability(0.1,1,5000) = %v, want %v", got, want)
	}

	// 即使单周期 p>=1，概率也被限制在 [0,1] 之间。
	if got := eventProbability(1000, 1, 1); got != 1 {
		t.Errorf("saturated rate = %v, want 1", got)
	}
}

func TestLockUserSerializes(t *testing.T) {
	// 在同一个用户锁下的并发增量不得产生数据竞争。
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
				counter++ // 读-改-写，没有锁的情况下会产生数据竞争
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
	// 不同的用户 ID 获取独立的锁（没有伪共享 / 死锁）。
	u1 := lockUser(100)
	u2 := lockUser(200) // 如果它们共享一个互斥锁，将永远阻塞
	u2()
	u1()
}
