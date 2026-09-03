package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunRequiresDatabaseURI(t *testing.T) {
	t.Setenv("DATABASE_URI", "")

	oldArgs := os.Args
	os.Args = []string{"gophermart"}
	defer func() { os.Args = oldArgs }()

	err := run()
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URI is required") {
		t.Fatalf("err = %v", err)
	}
}
