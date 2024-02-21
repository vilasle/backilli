package process

import (
	"os"

	"github.com/pkg/errors"
	"github.com/vilasle/backilli/internal/database"
	"github.com/vilasle/backilli/internal/entity"
	"github.com/vilasle/backilli/internal/period"
	env "github.com/vilasle/backilli/pkg/fs/environment"
	"github.com/vilasle/backilli/pkg/fs/manager"
	"gopkg.in/yaml.v2"
)

const (
	LocalVolume         = "local"
	SMBVolume           = "smb"
	YandexStorageVolume = "yandex.storage"
)

const (
	dbmsPostgresql = "pgsql"
)

type Env map[string]string

type Catalogs struct {
	Assets     string `yaml:"assets"`
	Transitory string `yaml:"transitory"`
}

type VolumeConfig struct {
	Id         string `yaml:"id"`
	Type       string `yaml:"type"`
	Address    string `yaml:"address"`
	Domain     string `yaml:"domain"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	MountPoint string `yaml:"mountpoint"`
	KeyId      string `yaml:"key_id"`
	KeySecret  string `yaml:"key_secret"`
	BucketName string `yaml:"bucket_name"`
	Root       string `yaml:"root"`
	Region     string `yaml:"region"`
}

type Tool struct {
	Postgresql struct {
		Frontend string `yaml:"psql"`
		Dumping  string `yaml:"dump"`
	} `yaml:"postgresql"`
	Compessing struct {
		Zip string `yaml:"7z"`
	} `yaml:"compessing"`
}

type Task struct {
	Id          string       `yaml:"id"`
	Type        string       `yaml:"type"`
	PartOfMonth string       `yaml:"part_of_month"`
	Repeat      []int        `yaml:"repeat"`
	Databases   []Database   `yaml:"dbs"`
	Files       []FileConfig `yaml:"files"`
	Compress    bool         `yaml:"compress"`
	Volumes     []string     `yaml:"volumes"`
	KeepCopies  int          `yaml:"keepCopies"`
}

type FileConfig struct {
	Path          string `yaml:"path"`
	IncludeRegexp string `yaml:"include_regexp"`
	ExcludeRegexp string `yaml:"exclude_regexp"`
}

type ProcessConfig struct {
	Env              `yaml:"enviroments"`
	DatabaseManagers `yaml:"dbms_managers"`
	Catalogs         `yaml:"catalogs"`
	Volumes          []VolumeConfig `yaml:"volumes"`
	ExternalTools    Tool           `yaml:"external_tool"`
	Tasks            []Task         `yaml:"tasks"`
	Events           `yaml:"events"`
}

type DatabaseManager struct {
	Name      string `yaml:"name"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	Interface string `yaml:"interface"`
}

type DatabaseManagers []DatabaseManager

func (m DatabaseManagers) GetAsSliceOfMaps() map[string]map[string]any {
	res := make(map[string]map[string]any)
	for _, v := range m {
		res[v.Name] = map[string]any{ 
			"host": v.Host, 
			"port": v.Port, 
			"user": v.User, 
			"password": v.Password, 
			"dbInterface": v.Interface,
		}
	}
	return res
}

type Events struct {
	BeforeStart  []string `yaml:"beforeStart"`
	BeforeFinish []string `yaml:"beforeFinish"`
}

type Database struct {
	Name    string `yaml:"name"`
	Manager string `yaml:"manager"`
}

func NewProcessConfig(path string) (ProcessConfig, error) {
	pc := ProcessConfig{}

	file, err := os.Open(path)
	if err != nil {
		return pc, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	err = d.Decode(&pc)

	return pc, err
}

func (pc ProcessConfig) SetEnviroment() error {
	for k, v := range pc.Env {
		if err := env.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (pc *ProcessConfig) PGDump() string {
	return pc.ExternalTools.Postgresql.Dumping
}

func (pc *ProcessConfig) Psql() string {
	return pc.ExternalTools.Postgresql.Frontend
}

func (pc *ProcessConfig) Compressing() string {
	return pc.ExternalTools.Compessing.Zip
}

func CreateBuilderConfigFromTask(
	task Task,
	volumes []manager.ManagerAtomic,
	rule period.PeriodRule, 
	dbmanagers database.Managers) ([]entity.BuilderConfig, error) {

	cfgs := make([]entity.BuilderConfig, 0)

	main := entity.BuilderConfig{
		Id:         task.Id,
		FsManagers: volumes,
		Compress:   task.Compress,
		Keep:       task.KeepCopies,
	}

	for _, db := range task.Databases {
		c := main

		switch db.Manager {
		case dbmsPostgresql:
			c.Type = entity.POSTGRESQL
		default:
			return nil, errors.Errorf("unknown type of manager '%s'", db.Manager)
		}
		c.Database = db.Name
		c.PeriodRule = rule
		
		if v, ok := dbmanagers[db.Manager]; ok {
			c.DatabaseManager = v
		} else {
			return nil, errors.Errorf("does not define database manager in task %v", db )
		}

		cfgs = append(cfgs, c)
	}

	for _, f := range task.Files {
		c := main
		c.Type = entity.FILE
		c.FilePath = f.Path
		c.PeriodRule = rule
		c.IncludeRegexp = f.IncludeRegexp
		c.ExcludeRegexp = f.ExcludeRegexp
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}
