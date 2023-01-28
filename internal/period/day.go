package period

import (
	"time"
)

const(
	DAILY = "daily"
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

type WeekdaysRule struct {
	ExcludeDays []int
}

func NewWeekdaysRule(days []int) WeekdaysRule {
	wd := map[int]int{
		Mon: 0,
		Tue: 0,
		Wed: 0,
		Thu: 0,
		Fri: 0,
		Sat: 0,
		Sun: 0,
	}
	
	for _, v := range days {
		wd[v]++
	}

	excludeDays := make([]int, 0)
	for k, v := range wd {
		if v == 0 {
			excludeDays = append(excludeDays, k)
		}
	} 

	return WeekdaysRule{
		ExcludeDays: excludeDays,
	}
}

func (r WeekdaysRule) NeedToExecute(t time.Time) bool {
	weekday := int(t.Weekday())
	return !r.intoExcluding(weekday)
}

func (r WeekdaysRule) intoExcluding(wd int) bool {
	for _, v := range r.ExcludeDays {
		if v == wd {
			return true
		}
	}
	return false
}
