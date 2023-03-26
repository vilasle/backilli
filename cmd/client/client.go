package main

// FIXME errors with smb. I think it is connected with netword quality, need to try to do write thought attemps and checking connection
// TODO save plan to restore
// TODO need to create json configuration for each other task and each other day. With help this check existing copies and use this for restoring data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	cfg "github.com/vilamslep/backilli/internal/config"
	ps "github.com/vilamslep/backilli/internal/process"
	"github.com/vilamslep/backilli/internal/report"
	"github.com/vilamslep/backilli/pkg/logger"
)

var (
	configPath string
	showHelp   bool
	loggerPath string

	//errors
	configErr error = errors.New("does not define config file")
)

func main() {

	var conf cfg.ProcessConfig
	var proc *ps.Process

	setCliAgrs()

	if showHelp {
		pflag.Usage()
		return
	}

	logger.InitLogger(loggerPath)

	if err := checkArgs(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		logger.Errorf("config file '%s' is not exists", configPath)
		os.Exit(2)
	}
	t := time.Now()
	logger.Infof("init procces from %s", configPath)
	{
		conf, err = cfg.NewProcessConfig(configPath)
		if err != nil {
			logger.Error("could not read config file", err)
			os.Exit(3)
		}

		logger.Debugf("config was read. Result = %v", conf)

		proc, err = ps.InitProcess(conf)
		if err != nil {
			logger.Error("could not init process", err)
			os.Exit(4)
		}
	}
	logger.Info("run process")
	{
		proc.Run()
		if err := proc.Close(); err != nil {
			logger.Error("could not finish process", err)
			os.Exit(5)
		}

		r := report.InitReports(proc)
		if content, err := json.Marshal(r); err != nil {
			logger.Error(err)
		} else {
			fd, err := os.Create(fmt.Sprintf("report_%s.json", t.Format("02-01-2006")))
			if err != nil {
				logger.Error(err)
			}
			if _, err := fd.Write(content); err != nil {
				logger.Error(err)
			}
		}
	}
}

func checkArgs() error {
	if configPath == "" {
		return configErr
	}
	return nil
}

func setCliAgrs() {
	pflag.BoolVarP(&showHelp, "help", "",
		false,
		"Print help message")
	pflag.StringVarP(&configPath, "config", "c",
		"",
		"Config file with setting of client. This supports YML file only")
	pflag.StringVarP(&loggerPath, "log", "l",
		"",
		"Path to log file. If is not filled then log will write to stdout")
	pflag.Parse()
}

// Short map and plans
// illi [operation] {init | run | check} [subject] [--config | -c ]  <string> [--env | -e] <string> [<Additional operation args>]
// Operation:
//	init
//		- subject:
//			- config  (run CLI interface generation config file )
//			- env	  (run CLI interface generation env file )
//  run
//		#TODO Don't want to do it
//		- subject:
//			- interface - running web interface for setting and monitoring (does not work)
//				- additional_agrs:
//					--http	  - listen to port ( default:1780)
//			- backup    - running backup
//
//	check
//		- subject:
//			- all ( load all setting and check access, e.g access to disk or postgres)
//illi interface --config file.conf --env file.env --http 8080
//	comment: run http server which will listen port 8080 and after finishing setting it write all setting to file.conf and file.env
//
//illi run backup --config file.conf --env file.env
//	comment: run backup process with using setting from file.conf and file.env
//
//illi init config --config file.conf
//	comment: run cli and generate config and save it to file.conf
//
//illi init env --env file.conf
//	comment: run cli and generate enviroment vars and save it to file.env
//
