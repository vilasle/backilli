package process

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vilamslep/backilli/internal/action/dump/postgresql"
	cfg "github.com/vilamslep/backilli/internal/config"
	"github.com/vilamslep/backilli/internal/entity"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/internal/tool/compress"
	"github.com/vilamslep/backilli/pkg/fs/manager"
	"github.com/vilamslep/backilli/pkg/fs/unit"
	"github.com/vilamslep/backilli/pkg/logger"
)

type Volume map[string]manager.ManagerAtomic

type Process struct {
	t        time.Time
	catalogs cfg.Catalogs
	entityes []entity.Entity
	volumes  Volume
}

func NewProcess() (*Process, error) {
	return nil, nil
}

func (ps *Process) Entityes() []entity.Entity {
	return ps.entityes
}

func InitProcess(conf cfg.ProcessConfig) (*Process, error) {
	process := Process{}

	logger.Debug("load enviroment vars")
	{
		if err := conf.SetEnviroment(); err != nil {
			return nil, errors.Wrap(err, "could not set enviroment vars")
		}
	}

	postgresql.PG_DUMP = conf.PGDump()
	postgresql.PSQL = conf.Psql()
	compress.Compressing = conf.Compressing()

	process.catalogs = conf.Catalogs

	logger.Debug("prepare config for initing volumes")
	{
		cfgs, err := convertConfigForFSManagers(conf.Volumes)
		if err != nil {
			return nil, err
		}

		logger.Debug("init volumes")
		{
			if ms, err := manager.InitManagersFromConfigs(cfgs); err == nil {
				process.volumes = ms
			} else {
				return nil, errors.Wrap(err, "could not init volumes")
			}
		}
	}

	logger.Debug("init tasks")
	{
		if err := process.setEntityFromTask(conf.Tasks); err != nil {
			return nil, errors.Wrap(err, "could not init tasks")
		}
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

func (ps *Process) Run() {
	t := time.Now()
	ps.t = t
	s := entity.EntitySetting{Tempdir: ps.catalogs.Transitory}
	for _, ent := range ps.entityes {
		logger.Info("checking period rules")
		{
			if !ent.CheckPeriodRules(t) {
				continue
			}
		}

		logger.Infof("run %v backup", ent)
		{
			if ent.Backup(s, t); ent.Err() != nil {
				logger.Info("an error occurred during backup", ent.Err())
				continue
			} else {
				logger.Infof("entity was finished success %v", ent)
			}
		}
	}
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
		if v.Type == period.DAILY {
			rule.Day = period.NewWeekdaysRule(v.Repeat)
		} else if v.Type == period.MONTHLY {
			rule.Month = period.NewMonthRule(v.Repeat, period.PartOfMonth(v.PartOfMonth))
		} else {
			return errors.New("unexpected type of period")
		}
		volumes := make([]manager.ManagerAtomic, 0)
		for _, m := range v.Volumes {
			if v, ok := pc.volumes[m]; ok {
				volumes = append(volumes, v)
			}
		}

		cs := cfg.CreateBuilderConfigFromTask(v, volumes, rule)
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
