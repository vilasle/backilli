package process

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vilamslep/backilli/internal/action/dump/postgresql"
	pgdb "github.com/vilamslep/backilli/internal/database/postgresql"
	"github.com/vilamslep/backilli/internal/entity"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/internal/tool/compress"
	"github.com/vilamslep/backilli/pkg/fs/environment"
	"github.com/vilamslep/backilli/pkg/fs/manager"
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

type Volume map[string]manager.ManagerAtomic

type Process struct {
	catalogs Catalogs
	email    []Email
	entitys  []entity.Entity
	volumes  Volume
}

func NewProcess() (*Process, error) {
	return nil, nil
}

func (ps *Process) Run() {
	t := time.Now()
	s := entity.EntitySetting{Tempdir: ps.catalogs.Transitory}
	for _, ent := range ps.entitys {
		if !ent.CheckPeriodRules(t) {
			continue
		}
		if ent.Backup(s, t); ent.Err() != nil {
			log.Println(ent.Err())
			continue
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

func (pc *Process) setEntityFromTask(tasks []Task) error {
	for _, v := range tasks {
		rule := period.PeriodRule{}
		if v.Type == period.DAILY {
			rule.Day = period.NewWeekdaysRule(v.Repeat)
		} else if v.Type == period.MONTHLY {
			rule.Month = period.NewMonthRule(v.Repeat, period.PartOfMonth(v.PartOfMonth))
		} else {
			return errors.New("unexpected type of period")
		}

		if len(v.Files) > 0 {
			pc.filesBackup(v, rule)
		}

		if len(v.PgDatabases) > 0 {
			pc.pgBackup(v, rule)
		}
	}
	return nil
}

func (pc *Process) pgBackup(t Task, rule period.PeriodRule) {
	for _, r := range t.PgDatabases {
		e := entity.PostgresEntity{
			Id:         t.Id,
			Database:   r,
			Compress:   t.Compress,
			PeriodRule: rule,
		}
		e.ConnectionConfig = pgdb.ConnectionConfig{
			User: environment.Get("PGUSER"),
			Password: environment.Get("PGPASSWORD"),
			SSlMode: false,			
		}

		for _, m := range t.Volumes {
			if v, ok := pc.volumes[m]; ok {
				e.FileManagers = append(e.FileManagers, v)
			}
		}
		pc.entitys = append(pc.entitys, e)
	}
}

func (pc *Process) filesBackup(t Task, rule period.PeriodRule) {
	for _, r := range t.Files {
		e := entity.FileEntity{
			Id:         t.Id,
			FilePath:   r.Path,
			Compress:   t.Compress,
			PeriodRule: rule,
		}
		if len(r.IncludeRegexp) > 0 {
			if re, err := regexp.Compile(r.IncludeRegexp); err == nil {
				e.IncludeRegexp = re
			} else {
				//TODO log
			}
		}

		if len(r.ExcludeRegexp) > 0 {
			if re, err := regexp.Compile(r.ExcludeRegexp); err == nil {
				e.ExcludeRegexp = re
			} else {
				//TODO log
			}
		}

		for _, m := range t.Volumes {
			if v, ok := pc.volumes[m]; ok {
				e.FileManagers = append(e.FileManagers, v)
			}
		}
		pc.entitys = append(pc.entitys, e)
	}
}

func InitProcess(conf ProcessConfig) (*Process, error) {
	process := Process{}

	if err := conf.SetEnviroment(); err != nil {
		return nil, err
	}

	postgresql.PG_DUMP = conf.PGDump()
	postgresql.PSQL = conf.Psql()
	compress.Compressing = conf.Compressing()

	process.catalogs = conf.Catalogs
	process.email = conf.Emails

	cfgs, err := convertConfigForFSManagers(conf.Volumes)
	if err != nil {
		return nil, err
	}

	if ms, err := manager.InitManagersFromConfigs(cfgs); err == nil {
		process.volumes = ms
	} else {
		return nil, err
	}

	if err := process.setEntityFromTask(conf.Tasks); err != nil {
		return nil, err
	}

	return &process, nil
}

func convertConfigForFSManagers(ms []VolumeConfig) ([]unit.ClientConfig, error) {
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
		if v.Type == SMBVolume {
			socket := strings.Split(v.Address, ":")
			c.Host = socket[0]
			if len(socket) != 2 {
				c.Port = 445
			} else {
				if p, err := strconv.Atoi(socket[1]); err == nil {
					c.Port = p
				} else {
					return nil, err
				}
			}
		}

		switch v.Type {
		case LocalVolume:
			c.Type = manager.LOCAL
		case SMBVolume:
			c.Type = manager.SMB
		case YandexStorageVolume:
			c.Type = manager.YANDEX
		default:
			return nil, errors.New("unexpected type of volume")
		}

		res = append(res, c)
	}
	return res, nil
}
