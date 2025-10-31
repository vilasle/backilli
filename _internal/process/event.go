package process

type eventsManager struct {
	beforeStart  []string
	beforeFinish []string
}

func (ps *Process) beforeStart() error {
	return ps.events.BeforeStart()
}

func (ps *Process) beforeFinish() error {
	return ps.events.BeforeFinish()
}