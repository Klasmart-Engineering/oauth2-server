package test

import (
	"os"
	"testing"
)

func Chdir(t *testing.T, dir string) func() {
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("err: %v", err)
	}

	return func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("err: %v", err)
		}
	}
}
