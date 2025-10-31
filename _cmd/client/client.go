package main

//FIXME errors with smb. I think it is connected with netword quality, need to try to do write thought attemps and checking connection
//TODO save plan to restore
//TODO need to create json configuration for each other task and each other day. With help this check existing copies and use this for restoring data

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"errors"

	s "github.com/vilasle/backilli/internal/config"
	p "github.com/vilasle/backilli/internal/process"
	"github.com/vilasle/backilli/internal/report"
	"github.com/vilasle/backilli/pkg/logger"
)

func main() {
	setting := cliSetting{}
	setting.Init()
	if err := setting.loadSettings(); err != nil {
		fmt.Printf("could not init settings from cli args by reason %v", err)
		setting.helpMessage()
		os.Exit(1)
	}

	if setting.showHelpIfIsNecessary() {
		return
	}
	defer setting.close()

	startApplication(setting)
}

/*
PROBLEMS

	FIXME 1. on linux system does not remove old backup on yandex cloud
*/
func startApplication(setting cliSetting) {
	var (
		conf s.ProcessConfig
		ps   *p.Process
		err  error
	)

	logger.Init(setting.environment, setting.output())

	logger.Info("initializing process", "path", setting.configPath)

	conf, err = s.NewProcessConfig(setting.configPath)
	if err != nil {
		logger.Error("could not read config file", "error", err)
		os.Exit(2)
	}

	logger.Debug("config was read", "config", conf)

	ps, err = p.InitProcess(conf)
	if err != nil {
		logger.Error("could not init process", err)
		os.Exit(3)
	}

	logger.Info("run process")

	//process can execute a long time
	//The base case using - it run on evening, after work and continue several hours
	//In executing time date can have changes because get time before running process
	//because I want to have correct report date
	t := time.Now()

	if err := ps.Execute(); err != nil {
		logger.Error("could not finish process", "error", err)
		os.Exit(4)
	}

	if err = saveReport(ps.Stat(), t); err != nil {
		logger.Error("saving report failed", err)
	}
}

func saveReport(stat *p.ProcessStat, reportDate time.Time) (err error) {
	var (
		buffer []byte
		fd     io.WriteCloser
	)

	reportFile := fmt.Sprintf("report_%s.json", reportDate.Format("02-01-2006"))

	if buffer, err = json.Marshal(report.InitReports(stat)); err != nil {
		return errors.Join(err, fmt.Errorf("marshalling report"))
	}

	if fd, err = os.Create(reportFile); err != nil {
		return errors.Join(err, fmt.Errorf("creating report file"))
	}
	defer fd.Close()

	if _, err = fd.Write(buffer); err != nil {
		return errors.Join(err, fmt.Errorf("could not write report to '%s'", reportFile))
	}

	return
}
