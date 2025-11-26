package models

type Command string

const (
	Start         Command = "/start"
	SetPartner    Command = "/set_partner"
	DeletePartner Command = "/delete_partner"
	Cancel        Command = "/cancel"
	Delete        Command = "/delete"
	AddCompliment Command = "/add_compliment"
)
