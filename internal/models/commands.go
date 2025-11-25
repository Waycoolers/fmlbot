package models

type Command string

const (
	Start         Command = "/start"
	Setpartner    Command = "/set_partner"
	Cancel        Command = "/cancel"
	Delete        Command = "/delete"
	AddCompliment Command = "/add_compliment"
)
