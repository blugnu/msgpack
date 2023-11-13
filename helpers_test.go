package msgpack

import (
	"errors"
	"testing"
)

func testError(t *testing.T, wanted, got error) {
	t.Helper()

	if wanted == nil && got != nil {
		t.Errorf("\nunexpected error: %#v\n\n", got)
	} else if !errors.Is(got, wanted) {
		t.Errorf("\nwanted %#v\ngot    %#v\n\n", wanted, got)
	}
}

func testPanic(t *testing.T, wanted error) {
	t.Helper()

	got := recover()
	if wanted == nil && got == nil {
		return
	}

	switch {
	case wanted == nil:
		t.Errorf("\nunexpected panic: %v\n\n", got)

	case got == nil:
		t.Errorf("\nwanted panic: %v\n\n", wanted)

	default:
		got, gotError := got.(error)
		if !gotError || !errors.Is(got, wanted) {
			t.Errorf("\nwanted panic: %v\ngot panic   : %#v\n\n", wanted, got)
		}
	}
}
