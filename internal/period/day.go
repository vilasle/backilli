package period

import (
	"time"
)

const(
	Mon = 1
	Tue = 2
	Wed = 3
	Thu = 4
	Fri = 5
	Sat = 6
	Sun = 7
)

type DaysOfWeekRule struct {
	ExcludeDays []int
}

func (r DaysOfWeekRule) NeedToExecute(t time.Time) bool {
	weekday := int(t.Weekday())
	return !r.intoExcluding(weekday)
}

func (r DaysOfWeekRule) intoExcluding(wd int) bool {
	for _, v := range r.ExcludeDays {
		if v == wd {
			return true
		}
	}
	return false
}
