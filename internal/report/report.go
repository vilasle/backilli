package report

import (
	"time"

	ps "github.com/vilasle/backilli/internal/process"
)

type Reports []Report

type Report struct {
	Date       time.Time `json:"date"`
	Name       string    `json:"id"`
	OID        string    `json:"name"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"finishTime"`
	SourceSize int64     `json:"sourceSize"`
	BackupSize int64     `json:"backupSize"`
	Paths      []string  `json:"paths"`
	Details    string    `json:"details"`
}

func InitReports(process *ps.Process) Reports {
	rps := make(Reports, 0)
	stat := process.Stat()
	for _, e := range stat.Entityes() {
		r := Report{
			Date:       stat.Date,
			Name:       e.Id(),
			OID:        e.OID(),
			Status:     e.Status(),
			StartTime:  e.StartTime(),
			EndTime:    e.EndTime(),
			SourceSize: e.EntitySize(),
			BackupSize: e.BackupSize(),
			Paths:      e.BackupPaths(),
		}
		if e.Err() != nil {
			r.Details = e.Err().Error()
		}
		rps = append(rps, r)
	}
	return rps
}
