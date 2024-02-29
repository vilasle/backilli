package process

import (
	"bytes"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	entityes     []entity.Entity
	volumes      Volume
	events       eventsManager
}

func (ps *Process) Entityes() []entity.Entity {
	return ps.entityes
}

func InitProcess(conf cfg.ProcessConfig) (*Process, error) {
	process := Process{}

	logger.Debug("loading enviroment vars")
	if err := conf.SetEnviroment(); err != nil {
		return nil, errors.Wrap(err, "could not set enviroment vars")
	}

	postgresql.PG_DUMP = conf.PGDump()
	postgresql.PSQL = conf.Psql()
	compress.Compressing = conf.Compressing()

	process.catalogs = conf.Catalogs

	logger.Debug("preparing config for initing volumes")
	cfgs, err := convertConfigForFSManagers(conf.Volumes)
	if err != nil {
		return nil, err
	}

	logger.Debug("initting volumes")
	if ms, err := manager.InitManagersFromConfigs(cfgs); err == nil {
		process.volumes = ms
	} else {
		return nil, errors.Wrap(err, "could not init volumes")
	}

	logger.Debug("initting database managers")
	if len(conf.DatabaseManagers) > 0 {
		if md, err := database.InitManagersFromConfig(conf.DatabaseManagers); err == nil {
			process.dbmsManagers = md
		} else {
			return nil, errors.Wrap(err, "could not init database managers")
		}
	} else {
		logger.Debug("there are not database managers in config")
	}

	logger.Debug("init tasks")
	if err := process.setEntityFromTask(conf.Tasks); err != nil {
		return nil, errors.Wrap(err, "could not init tasks")
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

	for _, ent := range ps.entityes {
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
		err := errors.New("closing error")
		for _, ferr := range e {
			err = errors.Wrap(err, ferr.Error())
		}
		return err
	}
	return nil
}

func (p *Process) Stat() *ProcessStat {
	stat := &ProcessStat{
		ps:       p,
		Date:     p.t,
		entityes: p.entityes,
	}
	return stat
}
