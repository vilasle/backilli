package process

import (
	"time"

	"github.com/vilasle/backilli/internal/entity"
)

type ProcessStat struct {
	ps       *Process
	Date     time.Time
	entityes []entity.Entity
}

func (stat *ProcessStat) Entityes() []entity.EntityInfo {
	es := make([]entity.EntityInfo, 0, len(stat.entityes))
	for i := range stat.entityes {
		if v, ok := stat.entityes[i].(entity.EntityInfo); ok {
			if v.Status() != "" {
				es = append(es, v)
			}
		}
	}
	return es
}
