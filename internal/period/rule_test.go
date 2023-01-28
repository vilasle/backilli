package period

import (
	"testing"
	"time"
)

func TestMonthRule(t *testing.T) {
	//beggining of jun
	date1 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.Now().Location())
	//middle of jun
	date2 := time.Date(2022, 1, 15, 0, 0, 0, 0, time.Now().Location())
	//finish of feb
	date3 := time.Date(2022, 2, 28, 0, 0, 0, 0, time.Now().Location())
	//other day of feb
	date4 := time.Date(2022, 2, 14, 0, 0, 0, 0, time.Now().Location())
	//beggining of march
	date5 := time.Date(2022, 3, 1, 0, 0, 0, 0, time.Now().Location())

	rule1 := MonthRule{
		PartOfMonth:  BEGGINING,
		ExcludeMonth: []int{},
	}

	rule2 := MonthRule{
		PartOfMonth:  MIDDLE,
		ExcludeMonth: []int{},
	}

	rule3 := MonthRule{
		PartOfMonth:  FINISH,
		ExcludeMonth: []int{},
	}

	rule4 := MonthRule{
		PartOfMonth:  BEGGINING,
		ExcludeMonth: []int{},
	}

	rule5 := MonthRule{
		PartOfMonth:  BEGGINING,
		ExcludeMonth: []int{MAR},
	}

	if ok := rule1.NeedToExecute(date1); !ok {
		t.Fatal("It's begginig of January. Should be True ")
	}

	if ok := rule2.NeedToExecute(date2); !ok {
		t.Fatal("It's middle of January. Should be True ")
	}

	if ok := rule3.NeedToExecute(date3); !ok {
		t.Fatal("It's finish of February. Should be True ")
	}

	if ok := rule4.NeedToExecute(date4); ok {
		t.Fatal("It's other of February. Should be False ")
	}

	if ok := rule5.NeedToExecute(date5); ok {
		t.Fatal("It's begginig of March. Should be False ")
	}
}

func TestWeekdaysRule(t *testing.T) {
	//Sunday
	date1 := time.Date(2022, 12, 25, 0, 0, 0, 0, time.Now().Location())
	//Monday
	date2 := time.Date(2022, 12, 26, 0, 0, 0, 0, time.Now().Location())

	rule := WeekdaysRule{
		ExcludeDays: []int{Sun, Sat},
	}

	if ok := rule.NeedToExecute(date1); ok {
		t.Log("It's Sunday. Shoud be False")
	}

	if ok := rule.NeedToExecute(date2); !ok {
		t.Log("It's Monday. Shoud be True")
	}
}
