package domain

type State string

const (
	Empty                       State = ""
	AwaitingPartner             State = "awaiting_partner"
	AwaitingCompliment          State = "awaiting_compliment"
	AwaitingComplimentFrequency State = "awaiting_compliment_frequency"
)
