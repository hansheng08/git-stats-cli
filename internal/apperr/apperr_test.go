package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hansheng08/git-stats-cli/internal/apperr"
)

var errBase = errors.New("base failure")

func TestUsage_WrapsAndClassifies(t *testing.T) {
	w := apperr.Usage(errBase)
	if !errors.Is(w, errBase) {
		t.Error("errors.Is(Usage(err), err) = false, want true")
	}
	if got, want := w.Error(), errBase.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !apperr.IsUsage(w) {
		t.Error("IsUsage(Usage(err)) = false, want true")
	}
}

func TestRuntime_WrapsAndClassifies(t *testing.T) {
	w := apperr.Runtime(errBase)
	if !errors.Is(w, errBase) {
		t.Error("errors.Is(Runtime(err), err) = false, want true")
	}
	if got, want := w.Error(), errBase.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if apperr.IsUsage(w) {
		t.Error("IsUsage(Runtime(err)) = true, want false")
	}
}

func TestIsUsage_ThroughFmtErrorfWrap(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"usage wrapped by Errorf", fmt.Errorf("context: %w", apperr.Usage(errBase)), true},
		{"usage wrapped twice", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", apperr.Usage(errBase))), true},
		{"runtime wrapped by Errorf", fmt.Errorf("context: %w", apperr.Runtime(errBase)), false},
		{"plain error", errBase, false},
		{"plain error wrapped by Errorf", fmt.Errorf("context: %w", errBase), false},
		{"usage inside runtime", apperr.Runtime(apperr.Usage(errBase)), true},
		{"runtime inside usage", apperr.Usage(apperr.Runtime(errBase)), true},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apperr.IsUsage(tt.err); got != tt.want {
				t.Errorf("IsUsage(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestNilPassThrough(t *testing.T) {
	if got := apperr.Usage(nil); got != nil {
		t.Errorf("Usage(nil) = %v, want nil", got)
	}
	if got := apperr.Runtime(nil); got != nil {
		t.Errorf("Runtime(nil) = %v, want nil", got)
	}
}
