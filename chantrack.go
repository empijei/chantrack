// Package chantrack allows to keep track of channels buffers.
package chantrack

import (
	"fmt"
	"iter"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"weak"
)

var (
	track typedMap[uint64, sampler]
	id    atomic.Uint64
)

// Make should be used instead of the builtin make to construct channels to track.
func Make[T any](size int) chan T {
	// NOTE: consider adding logging here if size>$THRESHOLD

	ch := make(chan T, size)
	if size == 0 {
		return ch
	}
	callerLoc := caller(1)
	add(ch, callerLoc)
	return ch
}

// Report contains info about a channel.
type Report struct {
	// MakeCallLocation is the callsite of Make that caused the channel to be
	// constructed and tracked.
	MakeCallLocation string
	Len, Cap         int
}

// Sample takes a sample of the currently tracked channels.
func Sample() []Report {
	var (
		gcd []uint64
		res []Report
	)
	for id, smpl := range track.Iter() {
		ln, cp, ok := smpl.read()
		if !ok {
			gcd = append(gcd, id)
		}
		res = append(res, Report{smpl.callerLoc, ln, cp})
	}

	slices.SortFunc(res, func(a, b Report) int {
		switch {
		case a.MakeCallLocation < b.MakeCallLocation:
			return -1
		case a.MakeCallLocation > b.MakeCallLocation:
			return 1
		default:
			return 0
		}
	})

	go func() { // Garbage collection
		for _, id := range gcd {
			track.Delete(id)
		}
	}()

	return res
}

func caller(depth int) string {
	pc, f, l, ok := runtime.Caller(1 + depth)
	if !ok {
		return "unable to find caller location"
	}
	fn := runtime.FuncForPC(pc)
	f = filepath.Base(f)
	n := fn.Name()
	return fmt.Sprintf("%s %s:%d", n, f, l)
}

type typedMap[K, V any] struct {
	s sync.Map
}

func (t *typedMap[K, V]) Store(k K, v V) {
	t.s.Store(k, v)
}

func (t *typedMap[K, V]) Delete(k K) {
	t.s.Delete(k)
}

func (t *typedMap[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		t.s.Range(func(key, value any) bool {
			return yield(key.(K), value.(V))
		})
	}
}

type sampler struct {
	callerLoc string
	read      func() (ln, cp int, ok bool)
}

func add[T any](ch chan T, callerLoc string) {
	wp := weak.Make(&ch)
	s := sampler{
		callerLoc: callerLoc,
		read: func() (ln, cp int, ok bool) {
			c := wp.Value()
			if c == nil {
				return 0, 0, false
			}
			return len(*c), cap(*c), true
		},
	}
	track.Store(id.Add(1), s)
}
