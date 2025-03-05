package chantrack

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTrack(t *testing.T) {
	defer reset()

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

	want := []Report{
		{
			MakeCallLocation: "github.com/empijei/chantrack.TestTrack chantrack_test.go:12",
			Len:              50,
			Cap:              100,
		},
		{
			MakeCallLocation: "github.com/empijei/chantrack.TestTrack chantrack_test.go:21",
			Cap:              20,
		},
		{
			MakeCallLocation: "github.com/empijei/chantrack.TestTrack chantrack_test.go:25",
			Len:              100,
			Cap:              100,
		},
	}

	if diff := cmp.Diff(want, Sample()); diff != "" {
		t.Error(diff)
	}
}
