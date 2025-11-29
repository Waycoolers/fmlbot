package models

type Command string

const (
	Start            Command = "/start"
	SetPartner       Command = "/set_partner"
	DeletePartner    Command = "/delete_partner"
	Cancel           Command = "/cancel"
	DeleteAccount    Command = "/delete_account"
	AddCompliment    Command = "/add_compliment"
	GetCompliments   Command = "/get_compliments"
	DeleteCompliment Command = "/delete_compliment"
)
