package process

import (
	"bytes"
	"runtime"
	"strings"
	"time"

	"errors"

	"github.com/vilasle/backilli/internal/action/dump/postgresql"
	cfg "github.com/vilasle/backilli/internal/config"
	"github.com/vilasle/backilli/internal/database"
	"github.com/vilasle/backilli/internal/entity"
	"github.com/vilasle/backilli/internal/tool/compress"
	"github.com/vilasle/backilli/pkg/fs/executing"
	"github.com/vilasle/backilli/pkg/fs/manager"
	"github.com/vilasle/backilli/pkg/logger"
)

type Volume map[string]manager.ManagerAtomic

type Process struct {
	t            time.Time
	catalogs     cfg.Catalogs
	dbmsManagers database.Managers
	entities     []entity.Entity
	volumes      Volume
	events       eventsManager
}

func (ps *Process) Entityes() []entity.Entity {
	return ps.entities
}

func InitProcess(conf cfg.ProcessConfig) (*Process, error) {
	process := Process{}

	logger.Debug("loading enviroment vars")
	if err := conf.SetEnvironment(); err != nil {
		return nil, errors.Join(err, errors.New("could not set environment vars"))
	}

	postgresql.PGDUMP = conf.PGDump()
	postgresql.PSQL = conf.Psql()
	compress.Compressing = conf.Compressing()

	process.catalogs = conf.Catalogs

	logger.Debug("preparing config for initialize volumes")
	configs, err := convertConfigForFSManagers(conf.Volumes)
	if err != nil {
		return nil, err
	}

	logger.Debug("init volumes")
	if ms, err := manager.InitManagersFromConfigs(configs); err == nil {
		process.volumes = ms
	} else {
		return nil, errors.Join(err, errors.New("could not init volumes"))
	}

	logger.Debug("init database managers")
	if len(conf.DatabaseManagers) > 0 {
		if md, err := database.InitManagersFromConfig(conf.DatabaseManagers); err == nil {
			process.dbmsManagers = md
		} else {
			return nil, errors.Join(err, errors.New("could not init database managers"))
		}
	} else {
		logger.Debug("there are not database managers in config")
	}

	logger.Debug("init tasks")
	if err := process.setEntityFromTask(conf.Tasks); err != nil {
		return nil, errors.Join(err, errors.New("could not init tasks"))
	}

	process.events = eventsManager{
		beforeStart:  conf.BeforeStart,
		beforeFinish: conf.BeforeFinish,
	}

	return &process, nil
}

func (ps *Process) Execute() error {
	ps.beforeStart()
	defer ps.beforeFinish()

	ps.t = time.Now()
	s := entity.EntitySetting{Tempdir: ps.catalogs.Transitory}

	for _, ent := range ps.entities {
		logger.Info("checking period rules", "task", ent)
		if !ent.CheckPeriodRules(ps.t) {
			logger.Info("checking did not executed. task will be skip", "task", ent)
			continue
		}

		st := time.Now()
		logger.Info("run backup", "task", ent)
		if ent.Backup(s, ps.t); ent.Err() != nil {
			logger.Error("an error occurred during backup",
				"task", ent,
				"error", ent.Err(),
				"time difference", time.Since(st).String())
		} else {
			logger.Info("entity was finished success", "task", ent)
		}
		runtime.GC()
	}
	return ps.Close()
}

func (mng eventsManager) BeforeStart() error {
	var (
		cmd, args []string
		app       string
		err       = errors.New("beforeFinish event has error")
		errs      = make([]error, 0, len(mng.beforeStart))
		stderr    bytes.Buffer
	)

	for _, c := range mng.beforeStart {
		if strings.Contains(c, " ") {
			cmd = strings.Split(c, " ")
			app, args = cmd[0], cmd[1:]
		} else {
			app = c
		}

		if err := executing.Execute(app, nil, &stderr, args...); err != nil {
			errs = append(errs, err)
		}
	}
	return joinErrors(errs, err)
}

func (mng eventsManager) BeforeFinish() error {
	var (
		cmd, args []string
		app       string
		err       = errors.New("beforeFinish event has error")
		errs      = make([]error, 0, len(mng.beforeStart))
		stderr    bytes.Buffer
	)

	for _, c := range mng.beforeFinish {
		if strings.Contains(c, " ") {
			cmd = strings.Split(c, " ")
			app, args = cmd[0], cmd[1:]
		} else {
			app = c
		}

		if err := executing.Execute(app, nil, &stderr, args...); err != nil {
			errs = append(errs, err)
		}
	}
	return joinErrors(errs, err)
}

func (ps *Process) Close() error {
	e := make([]error, 0)
	for _, v := range ps.volumes {
		if err := v.Close(); err != nil {
			e = append(e, err)
		}
	}
	if len(e) > 0 {
		e = append(e, errors.New("could not close volumes"))
	}
	return errors.Join(e...)
}

func (p *Process) Stat() *ProcessStat {
	stat := &ProcessStat{
		ps:       p,
		Date:     p.t,
		entities: p.entities,
	}
	return stat
}
