package errors

import (
	"errors"
	"testing"
)

func TestErrorsMatch(t *testing.T) {
	s := &StatusError{Code: 1}
	st := &StatusError{Code: 2}
	if errors.Is(s, st) {
		t.Errorf("error is not match: %+v -> %+v", s, st)
	}
	s.Code = 1
	st.Code = 1
	if !errors.Is(s, st) {
		t.Errorf("error is not match: %+v -> %+v", s, st)
	}

	s.Reason = "test_reason"
	s.Reason = "test_reason"

	if !errors.Is(s, st) {
		t.Errorf("error is not match: %+v -> %+v", s, st)
	}

	if Reason(s) != "test_reason" {
		t.Errorf("error is not match: %+v -> %+v", s, st)
	}
}