package report

import (
	"testing"

	cfg "github.com/vilamslep/backilli/internal/config"
	pc "github.com/vilamslep/backilli/internal/process"
	env "github.com/vilamslep/backilli/pkg/fs/environment"
	"github.com/vilamslep/backilli/pkg/logger"
)

func TestInitReport(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}
	logger.InitLogger("")
	path := env.Get("CONFIG")

	config, err := cfg.NewProcessConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	proc, err := pc.InitProcess(config)
	if err != nil {
		t.Fatal(err)
	}

	_ = InitReports(proc)
}
