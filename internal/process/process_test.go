package process

import (
	"testing"

	env "github.com/vilamslep/backilli/pkg/fs/environment"
)

func TestNewProcessConfig(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}

	path := env.Get("CONFIG")

	if _, err := NewProcessConfig(path); err != nil {
		t.Fatal(err)
	}
}
