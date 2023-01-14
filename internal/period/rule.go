package period

import (
	"time"
)

type Rule interface {
	NeedToExecute(time.Time) bool
}
