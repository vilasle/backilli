package period

import (
	"time"
)

type Rule interface {
	NeedToExecute(time.Time) bool
}

type PeriodRule struct {
	Day  Rule
	Month Rule
}