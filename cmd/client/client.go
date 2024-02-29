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

	"github.com/pkg/errors"
	s "github.com/vilasle/backilli/internal/config"
	p "github.com/vilasle/backilli/internal/process"
	"github.com/vilasle/backilli/internal/report"
	"github.com/vilasle/backilli/pkg/logger"
)

func main() {

	clis := cliSetting{}
	clis.setArgs()
	if err := clis.loadSettings(); err != nil {
		fmt.Println("could not init settings from cli args")
		fmt.Println(err)
		clis.helpMessage()
		os.Exit(1)
	}

	if clis.showHelpIfIsNecessary() {
		return
	}
	defer clis.close()

	startApplication(clis)

	logger.Init(clis.enviroment, clis.output())

	startApplication(clis)
}

func startApplication(clis cliSetting) {
	var (
		conf s.ProcessConfig
		ps   *p.Process
		err  error
	)

	logger.Info("initing procces", "path", clis.configPath)

	conf, err = s.NewProcessConfig(clis.configPath)
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
		reportFormat     = "report_%s.json"
		reportTimeLayout = "02-01-2006"
		reportFile       = fmt.Sprintf(reportFormat, reportDate.Format(reportTimeLayout))
		c                []byte
		fd               io.WriteCloser
	)

	if c, err = json.Marshal(report.InitReports(stat)); err != nil {
		return errors.Wrap(err, "marshalling report")
	}

	if fd, err = os.Create(reportFile); err != nil {
		return errors.Wrap(err, "creating report file")
	}
	defer fd.Close()

	if _, err = fd.Write(c); err != nil {
		return errors.Wrapf(err, "could not write report to '%s'", reportFile)
	}

	return
}
