package period

import "time"

//Path of month
const (
	BEGGINING = 1
	MIDDLE    = 2
	FINISH    = 3
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
	PathOfMonth  int
	ExcludeMonth []int
}

func (r MonthRule) NeedToExecute(t time.Time) bool {

	month := int(t.Month())
	day := t.Day()

	if r.intoExcluding(month) {
		return false
	}

	if r.PathOfMonth == BEGGINING && day == 1 {
		return true
	}

	if r.PathOfMonth == MIDDLE && day == 15 {
		return true
	}

	if r.PathOfMonth == FINISH && isFinishOfMonth(t) {
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
