package process

import (
	"os"

	env "github.com/vilamslep/backilli/pkg/fs/environment"
	"gopkg.in/yaml.v2"
)

const (
	LocalVolume         = "local"
	SMBVolume           = "smb"
	YandexStorageVolume = "yandex.storage"
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
	Id          string   `yaml:"id"`
	Type        string   `yaml:"type"`
	PartOfMonth string   `yaml:"part_of_month"`
	Repeat      []int    `yaml:"repeat"`
	PgDatabases []string `yaml:"psql_dbs"`
	Files       []struct {
		Path          string `yaml:"path"`
		IncludeRegexp string `yaml:"include_regexp"`
		ExcludeRegexp string `yaml:"exclude_regexp"`
	} `yaml:"files"`
	Compress   bool     `yaml:"compress"`
	Volumes    []string `yaml:"volumes"`
	KeepCopies int      `yaml:"keepCopies"`
}

type ProcessConfig struct {
	Env           `yaml:"enviroments"`
	Catalogs      `yaml:"catalogs"`
	Volumes       []VolumeConfig `yaml:"volumes"`
	ExternalTools Tool           `yaml:"external_tool"`
	Tasks         []Task         `yaml:"tasks"`
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
