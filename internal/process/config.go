package process

import (
	"os"
	"gopkg.in/yaml.v2"
)

type Env map[string]string

type Catalogs map[string]string

type Volume struct {
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
}

type Email struct {
	Id          string   `yaml:"id"`
	User        string   `yaml:"user"`
	Password    string   `yaml:"password"`
	FromName    string   `yaml:"fromName"`
	SmtpAddress string   `yaml:"smtp"`
	Recivers    []string `yaml:"recivers"`
	Letter      struct {
		Subject string `yaml:"subject"`
	} `yaml:"letter"`
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
	Repeat      []int    `yaml:"repeat"`
	PgDatabases []string `yaml:"psql_dbs"`
	Files       []string `yaml:"files"`
	Compress    bool     `yaml:"compress"`
	Volumes     []string `yaml:"volumes"`
}

type ProcessConfig struct {
	Env           `yaml:"enviroments"`
	Catalogs      `yaml:"catalogs"`
	Volumes       []Volume `yaml:"volumes"`
	Emails        []Email  `yaml:"email"`
	ExternalTools Tool     `yaml:"external_tool"`
	Tasks         []Task   `yaml:"tasks"`
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
