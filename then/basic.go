package then

import (
	"errors"
	"strings"
	"testing"
)

// Equals compares two values, in some rare cases due to generic limitations
// you may have to use `reflect.DeepEquals` instead.
func Equals[T comparable](t testing.TB, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Logf("expected '%v' to equal '%v'", expected, actual)
		t.FailNow()
	}
}

// NotEquals compares two values are not equal
func NotEqual[T comparable](t testing.TB, expected, actual T) {
	t.Helper()

	if expected == actual {
		t.Logf("expected '%v' not to equal '%v'", expected, actual)
		t.FailNow()
	}
}

// SliceEquals compares two values, in some rare cases due to generic limitations
// you may have to use `reflect.DeepEquals` instead.
func SliceEquals[T comparable](t testing.TB, expected, actual []T) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Logf("expected len of '%v' to equal '%v'", expected, actual)
		t.FailNow()
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			t.Logf("expected '%v' to equal '%v' at index %d", expected[i], actual[i], i)
			t.FailNow()
		}
	}
}

// Nil compares a value to nil, in some cases you may need to do `Equals(t, value, nil)`
func Nil(t testing.TB, value any) {
	t.Helper()

	if value != nil {
		t.Logf("expected '%v' to be nil", value)
		t.FailNow()
	}
}

// NotNil compares a value is not nil.
func NotNil(t testing.TB, value any) {
	t.Helper()

	if value == nil {
		t.Logf("expected '%v' not to be nil", value)
		t.FailNow()
	}
}

// Err checks if our actual error is the expected error or wrapped in the expected error.
func Err(t testing.TB, expected, actual error) {
	t.Helper()

	if !errors.Is(actual, expected) {
		t.Logf("expected '%v' to be '%v'", expected, actual)
		t.FailNow()
	}
}

// Panic checks if our func would panic.
func Panic(t testing.TB, f func()) {
	defer func() {
		// we don't care what the value is, only that we had to recover
		_ = recover()
	}()

	f()
	t.Error("expected func to panic")
	t.FailNow()
}

// True checks if a value is true.
func True(t testing.TB, value bool) {
	t.Helper()

	if !value {
		t.Error("expected value to be true")
		t.FailNow()
	}
}

// False checks if a value is false.
func False(t testing.TB, value bool) {
	t.Helper()

	if value {
		t.Error("expected value to be false")
		t.FailNow()
	}
}

// Contains checks if our substring is contained in the full string
func Contains(t testing.TB, sub, full string) {
	t.Helper()

	if !strings.Contains(full, sub) {
		t.Logf("expected '%v' to be in '%v'", sub, full)
		t.FailNow()
	}
}
