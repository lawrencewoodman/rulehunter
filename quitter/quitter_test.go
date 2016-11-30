package quitter

import (
	"testing"
)

func TestQuit(t *testing.T) {
	quit := New()
	nums := make(chan int, 10)
	for i := 0; i < 10; i++ {
		quit.Add()
		go func(n int) {
			defer quit.Done()
			for {
				select {
				case <-quit.C:
					nums <- n
					return
				}
			}
		}(i)
	}

	quit.Quit()
	close(nums)

	sum := 0
	for n := range nums {
		sum += n
	}
	want := 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9
	if sum != want {
		t.Errorf("Quit: didn't wait for all processes to quit sum: %v", sum)
	}
}
