package chantrack

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestUnTrack(t *testing.T) {
	defer reset()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	func() {
		a := Make[int](100)
		for range 50 {
			a <- 42
		}

		b := Make[int](0)
		go func() { b <- 42 }()
		<-b

		c := Make[int](20)
		c <- 1
		<-c

		d := Make[int](100)
		for range 100 {
			d <- 42
		}
	}()
	for len(Sample()) > 0 {
		select {
		case <-ctx.Done():
			t.Fatalf("took too long to clear the channels")
		default:
			runtime.GC()
			runtime.Gosched()
		}
	}
}

func reset() {
	track = typedMap[uint64, sampler]{}
	id.Store(0)
}
