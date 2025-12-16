package domain

type State string

const (
	Empty                             State = ""
	AwaitingPartner                   State = "awaiting_partner"
	AwaitingCompliment                State = "awaiting_compliment"
	AwaitingComplimentFrequency       State = "awaiting_compliment_frequency"
	AwaitingTitleImportantDate        State = "awaiting_title_important_date"
	AwaitingDateImportantDate         State = "awaiting_date_important_date"
	AwaitingPartnerImportantDate      State = "awaiting_partner_important_date"
	AwaitingNotifyBeforeImportantDate State = "awaiting_notify_before_important_date"
	AwaitingConfirmImportantDate      State = "awaiting_confirm_important_date"
)
