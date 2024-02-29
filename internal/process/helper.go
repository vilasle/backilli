package process

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	cfg "github.com/vilasle/backilli/internal/config"
	"github.com/vilasle/backilli/internal/entity"
	"github.com/vilasle/backilli/internal/period"
	"github.com/vilasle/backilli/pkg/fs/manager"
	"github.com/vilasle/backilli/pkg/fs/unit"
)

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
