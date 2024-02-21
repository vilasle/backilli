package process

import (
	"testing"

	cfg "github.com/vilasle/backilli/internal/config"
	env "github.com/vilasle/backilli/pkg/fs/environment"
	"github.com/vilasle/backilli/pkg/logger"
)

func TestNewProcessConfig(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}

	path := env.Get("CONFIG")

	if _, err := cfg.NewProcessConfig(path); err != nil {
		t.Fatal(err)
	}
}

func TestInitProcess(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}
	logger.InitLogger("local", nil)
	path := env.Get("CONFIG")

	conf, err := cfg.NewProcessConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := InitProcess(conf); err != nil {
		t.Fatal(err)
	}
}
