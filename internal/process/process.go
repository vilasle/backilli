package process

import (
	"bytes"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vilasle/backilli/internal/action/dump/postgresql"
	cfg "github.com/vilasle/backilli/internal/config"
	"github.com/vilasle/backilli/internal/database"
	"github.com/vilasle/backilli/internal/entity"
	"github.com/vilasle/backilli/internal/period"
	"github.com/vilasle/backilli/internal/tool/compress"
	"github.com/vilasle/backilli/pkg/fs/executing"
	"github.com/vilasle/backilli/pkg/fs/manager"
	"github.com/vilasle/backilli/pkg/fs/unit"
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

type eventsManager struct {
	beforeStart  []string
	beforeFinish []string
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

func joinErrors(errs []error, mainErr error) error {
	if len(errs) == 0 {
		return nil
	}
	serr := mainErr
	for _, err := range errs {
		serr = errors.Wrap(serr, err.Error())
	}
	return serr
}

// func NewProcess() (*Process, error) {
// 	return nil, nil
// }

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

func (p *Process) Stat() *ProcessStat {
	stat := &ProcessStat{
		ps:       p,
		Date:     p.t,
		entityes: p.entityes,
	}
	return stat
}

// events
func (ps *Process) beforeStart() error {
	return ps.events.BeforeStart()
}

func (ps *Process) beforeFinish() error {
	return ps.events.BeforeFinish()
}

func (ps *Process) Run() {
	ps.beforeStart()
	defer ps.beforeFinish()

	t := time.Now()
	ps.t = t
	s := entity.EntitySetting{Tempdir: ps.catalogs.Transitory}

	for _, ent := range ps.entityes {
		logger.Info("checking period rules", "task", ent)
		if !ent.CheckPeriodRules(t) {
			logger.Info("checking did not executed. task will be skip", "task", ent)
			continue
		}

		startTime := time.Now()
		logger.Info("run backup", "task", ent)
		if ent.Backup(s, t); ent.Err() != nil {
			logger.Error("an error occurred during backup",
				"task", ent,
				"error", ent.Err(),
				"time difference", diffWithNow(startTime).String())
		} else {
			logger.Info("entity was finished success", "task", ent)
		}
		runtime.GC()
	}
}

func diffWithNow(t1 time.Time) time.Duration {
	return time.Since(t1)
}

func (pc *Process) Close() error {
	e := make([]error, 0)
	for _, v := range pc.volumes {
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

func (pc *Process) setEntityFromTask(tasks []cfg.Task) error {
	for _, v := range tasks {
		rule := period.PeriodRule{}
		switch v.Type {
		case period.DAILY:
			rule.Day = period.NewWeekdaysRule(v.Repeat)
		case period.MONTHLY:
			rule.Month = period.NewMonthRule(v.Repeat, period.PartOfMonth(v.PartOfMonth))
		default:
			return errors.New("unexpected type of period")
		}

		volumes := make([]manager.ManagerAtomic, 0)
		for _, m := range v.Volumes {
			if v, ok := pc.volumes[m]; ok {
				volumes = append(volumes, v)
			}
		}

		cs, err := cfg.CreateBuilderConfigFromTask(v, volumes, rule, pc.dbmsManagers)
		if err != nil {
			return errors.Wrap(err, "there are errors on creation config tasks")
		}
		es, err := entity.CreateAllEntitys(cs)
		if err != nil {
			return errors.Wrapf(err, "could not create backup entity from config %v", cs)
		}
		pc.entityes = append(pc.entityes, es...)
	}
	return nil
}

func convertConfigForFSManagers(ms []cfg.VolumeConfig) ([]unit.ClientConfig, error) {
	res := make([]unit.ClientConfig, 0, len(ms))
	for _, v := range ms {
		c := unit.ClientConfig{}
		c.Id = v.Id
		c.BucketName = v.BucketName
		c.Domain = v.Domain
		c.MountPoint = v.MountPoint
		c.User = v.User
		c.Password = v.Password
		c.Root = v.Root
		c.KeyId = v.KeyId
		c.KeySecret = v.KeySecret
		c.Region = v.Region
		if v.Type == cfg.SMBVolume {
			socket := strings.Split(v.Address, ":")
			c.Host = socket[0]
			if len(socket) != 2 {
				c.Port = 445
			} else {
				if p, err := strconv.Atoi(socket[1]); err == nil {
					c.Port = p
				} else {
					return nil, errors.Wrapf(err, "does not convert smb socket %s to expected type", v.Address)
				}
			}
		}

		switch v.Type {
		case cfg.LocalVolume:
			c.Type = manager.LOCAL
		case cfg.SMBVolume:
			c.Type = manager.SMB
		case cfg.YandexStorageVolume:
			c.Type = manager.YANDEX
		default:
			return nil, errors.New("unexpected type of volume")
		}
		res = append(res, c)
	}
	return res, nil
}
