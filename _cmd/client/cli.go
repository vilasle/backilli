package main

import (
	"io"
	"os"

	"errors"

	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	errRequiredArguments    = errors.New("required arguments are not filled")
	errConfigFile           = errors.New("config file is not exists")
	errNotDefinedConfigFile = errors.New("does not define config file")
)

type cliSetting struct {
	configPath  string
	showHelp    bool
	loggerPath  string
	environment string
	logOut      io.WriteCloser
}

func (c *cliSetting) Init() {
	pflag.BoolVarP(&c.showHelp, "help", "",
		false,
		"Print help message")
	pflag.StringVarP(&c.configPath, "config", "c",
		"",
		"Config file with setting of client. This supports YAML file only")
	pflag.StringVarP(&c.loggerPath, "log", "l",
		"",
		"Path to log file. If is not filled then log will write to stdout")
	pflag.StringVarP(&c.environment, "env", "e",
		"local",
		"kind of environment running. Log level and format depend on this")
	pflag.Parse()
}

func (c cliSetting) output() io.WriteCloser {
	return c.logOut
}

func (c *cliSetting) loadSettings() (err error) {
	if err := c.checkArgs(); err != nil {
		return errors.Join(err, errRequiredArguments)
	}

	c.logOut = c.defineLogDestination()

	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		return errors.Join(err, errConfigFile)
	}
	return nil
}

func (c *cliSetting) checkArgs() error {
	if c.configPath == "" {
		return errNotDefinedConfigFile
	}
	return nil
}

func (c cliSetting) showHelpIfIsNecessary() bool {
	if c.showHelp {
		pflag.Usage()
	}
	return c.showHelp
}

func (c cliSetting) defineLogDestination() io.WriteCloser {
	if c.loggerPath == "" {
		return os.Stdout
	}

	return &lumberjack.Logger{
		Filename:   c.loggerPath,
		MaxSize:    10,
		MaxAge:     10,
		MaxBackups: 10,
		Compress:   true,
	}
}

func (c cliSetting) helpMessage() {
	pflag.Usage()
}

func (c *cliSetting) close() error {
	return c.logOut.Close()
}
