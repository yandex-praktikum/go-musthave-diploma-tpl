package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	l, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if l == nil {
		t.Fatal("New() returned nil logger")
	}
}
