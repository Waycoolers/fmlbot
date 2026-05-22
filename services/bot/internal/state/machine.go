package state

type Machine struct {
	userSteps map[int64]State
}

func New() *Machine {
	return &Machine{
		userSteps: make(map[int64]State),
	}
}

func (m *Machine) GetStep(userID int64) State {
	return m.userSteps[userID]
}

func (m *Machine) SetStep(userID int64, state State) {
	m.userSteps[userID] = state
}
