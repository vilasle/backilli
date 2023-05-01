package period

import "time"

const (
	MONTHLY = "monthly"
)

// Path of month
type PartOfMonth string

const (
	BEGGINING = "beggining"
	MIDDLE    = "middle"
	FINISH    = "finish"
)

const (
	JAN = 1
	FEB = 2
	MAR = 3
	APR = 4
	MAY = 5
	JUN = 6
	JUL = 7
	AUG = 8
	SEP = 9
	OCT = 10
	NOV = 11
	DEC = 12
)

type MonthRule struct {
	PartOfMonth  PartOfMonth
	ExcludeMonth []int
}

func NewMonthRule(month []int, partOfMonth PartOfMonth) MonthRule {
	m := map[int]int{
		JAN: 0,
		FEB: 0,
		MAR: 0,
		APR: 0,
		MAY: 0,
		JUN: 0,
		JUL: 0,
		AUG: 0,
		SEP: 0,
		OCT: 0,
		NOV: 0,
		DEC: 0,
	}

	for _, v := range month {
		m[v]++
	}

	excludeMonth := make([]int, 0)
	for k, v := range month {
		if v == 0 {
			excludeMonth = append(excludeMonth, k)
		}
	}

	return MonthRule{
		ExcludeMonth: excludeMonth,
		PartOfMonth:  partOfMonth,
	}
}

func (r MonthRule) NeedToExecute(t time.Time) bool {

	month := int(t.Month())
	day := t.Day()

	if r.intoExcluding(month) {
		return false
	}

	if r.PartOfMonth == BEGGINING && day == 1 {
		return true
	}

	if r.PartOfMonth == MIDDLE && day == 15 {
		return true
	}

	if r.PartOfMonth == FINISH && isFinishOfMonth(t) {
		return true
	}
	return false
}

func (r MonthRule) intoExcluding(month int) bool {
	for _, v := range r.ExcludeMonth {
		if v == month {
			return true
		}
	}
	return false
}

func isFinishOfMonth(t time.Time) bool {
	nextMonth := int(t.Month()) + 1
	year := t.Year()
	if int(t.Month()) == 12 {
		nextMonth = 1
		year++
	}
	//create new date which is start of next month and subtraction one day for knowing number of day which is the finish of current month
	t2 := time.Date(year, time.Month(nextMonth), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 0, -1)
	return t.Day() == t2.Day()
}
