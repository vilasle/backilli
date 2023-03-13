package main

import (
	"errors"
	"os"

	"github.com/spf13/pflag"
	ps "github.com/vilamslep/backilli/internal/process"
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

	setCliAgrs()

	if showHelp {
		pflag.Usage()
		return
	}

	logger.InitLogger((loggerPath == ""), loggerPath)

	if err := checkArgs(); err != nil {
		logger.Fatal(err)
	}

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		logger.Fatal("config file is not exists", configPath)
	}

	var (
		conf ps.ProcessConfig
		proc *ps.Process
	)
	logger.Info("init procces")
	{
		conf, err = ps.NewProcessConfig(configPath)
		if err != nil {
			logger.Fatal("could not read config file", configPath, err)
		}

		logger.Debugf("config was read. Result = %v", conf)

		proc, err = ps.InitProcess(conf)
		if err != nil {
			logger.Fatal("could not init process", err)
		}
	}
	logger.Info()

	logger.Info("run process")
	proc.Run()

	if err := proc.Close(); err != nil {
		logger.Fatal("could not finish process", err)
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
		"Confif file with setting of client. This supports YML file only")
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
