package state

type Machine struct {
	step State
}

func New() *Machine {
	return &Machine{
		step: Empty,
	}
}

func (m *Machine) GetStep() State {
	return m.step
}

func (m *Machine) SetStep(step State) {
	m.step = step
}
