package process

import (
	"time"

	"github.com/vilasle/backilli/internal/entity"
)

type ProcessStat struct {
	ps       *Process
	Date     time.Time
	entities []entity.Entity
}

func (stat *ProcessStat) Entities() []entity.EntityInfo {
	es := make([]entity.EntityInfo, 0, len(stat.entities))
	for i := range stat.entities {
		if v, ok := stat.entities[i].(entity.EntityInfo); ok {
			if v.Status() != "" {
				es = append(es, v)
			}
		}
	}
	return es
}
