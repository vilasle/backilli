package main

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type cliSetting struct {
	configPath string
	showHelp   bool
	loggerPath string
	enviroment string
	logOut     io.WriteCloser
}

func (c *cliSetting) setArgs() {
	pflag.BoolVarP(&c.showHelp, "help", "",
		false,
		"Print help message")
	pflag.StringVarP(&c.configPath, "config", "c",
		"",
		"Config file with setting of client. This supports YML file only")
	pflag.StringVarP(&c.loggerPath, "log", "l",
		"",
		"Path to log file. If is not filled then log will write to stdout")
	pflag.StringVarP(&c.enviroment, "env", "e",
		"local",
		"kind of enviroment running. Log level and format depend on this")
	pflag.Parse()
}

func (c cliSetting) output() io.WriteCloser {
	return c.logOut
}

func (c *cliSetting) loadSettings() error {
	if err := c.checkArgs(); err != nil {
		return errors.Wrap(err, "checking required arguments")
	}

	wrt, err := c.defineLogDestination()
	if err != nil {
		return errors.Wrap(err, "getting log file was failed")
	}

	c.logOut = wrt

	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		return errors.Wrap(err, "config file is not exists")
	}
	return nil
}

func (c *cliSetting) checkArgs() error {

	if c.configPath == "" {
		return errors.New("does not define config file")
	}

	return nil
}

func (c cliSetting) showHelpIfIsNecessary() bool {
	if c.showHelp {
		pflag.Usage()
	}
	return c.showHelp
}

func (c cliSetting) defineLogDestination() (io.WriteCloser, error) {
	var (
		wrt io.WriteCloser
		err error
	)
	if c.loggerPath == "" {
		return os.Stdout, nil
	}

	if wrt, err = os.OpenFile(c.loggerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm); err != nil {
		return os.Stdout, errors.Wrapf(err, "can not open file. Log will write to stdout")
	}
	return wrt, nil
}

func (c cliSetting) helpMessage() {
	pflag.Usage()
}

func (c *cliSetting) close() error {
	return c.logOut.Close()
}
