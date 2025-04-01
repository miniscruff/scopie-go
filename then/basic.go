package then

import (
	"errors"
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
