package main

//FIXME errors with smb. I think it is connected with netword quality, need to try to do write thought attemps and checking connection
//TODO save plan to restore
//TODO need to create json configuration for each other task and each other day. With help this check existing copies and use this for restoring data

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	cfg "github.com/vilasle/backilli/internal/config"
	ps "github.com/vilasle/backilli/internal/process"
	"github.com/vilasle/backilli/internal/report"
	"github.com/vilasle/backilli/pkg/logger"
)

var (
	configPath string
	showHelp   bool
	loggerPath string
	enviroment string

	//errors
	configErr        error = errors.New("does not define config file")
	reportFormat           = "report_%s.json"
	reportTimeLayout       = "02-01-2006"
)

func main() {
	var (
		conf cfg.ProcessConfig
		proc *ps.Process
	)

	setCliAgrs()
	if showHelp {
		pflag.Usage()
		return
	}

	logWriter, err := defineLogDestination()
	if err != nil {
		log.Println(err)
	}
	defer logWriter.Close()

	logger.Init(enviroment, logWriter)

	if err := checkArgs(); err != nil {
		logger.Error("checking required arguments", "error", err.Error())
		os.Exit(1)
	}

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		logger.Error("config is not exists", "path", configPath)
		os.Exit(2)
	}

	logger.Info("initing procces", "path", configPath)

	conf, err = cfg.NewProcessConfig(configPath)
	if err != nil {
		logger.Error("could not read config file", "error", err)
		os.Exit(3)
	}

	logger.Debug("config was read", "config", conf)

	proc, err = ps.InitProcess(conf)
	if err != nil {
		logger.Error("could not init process", err)
		os.Exit(4)
	}

	logger.Info("run process")

	//process can execute a long time
	//The base case using - it run on evening, after work and continue several hours
	//In executing time date can have changes because get time before running process
	//because I want to have correct report date
	t := time.Now()

	proc.Run()
	if err := proc.Close(); err != nil {
		logger.Error("could not finish process", "error", err)
		os.Exit(5)
	}

	r := report.InitReports(proc)
	content, err := json.Marshal(r)
	if err != nil {
		logger.Error("marshalling report", "error", err)
	}

	reportFile := fmt.Sprintf(reportFormat, t.Format(reportTimeLayout))
	fd, err := os.Create(reportFile)
	if err != nil {
		logger.Error("creating report file", "file", reportFile, "error", err)
	}

	if _, err := fd.Write(content); err != nil {
		logger.Error("writting report", "file", reportFile, "error", err)
	}

}

func defineLogDestination() (io.WriteCloser, error) {
	if loggerPath != "" {
		fd, err := os.OpenFile(loggerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			return os.Stdout, errors.Wrapf(err, "can not open file. Log will write to stdout")
		}
		return fd, nil
	} else {
		return os.Stdout, nil
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
	pflag.StringVarP(&enviroment, "env", "e",
		"local",
		"kind of enviroment running. Log level and format depend on this")
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
