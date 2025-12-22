package domain

type State string

const (
	Empty                                 State = ""
	AwaitingPartner                       State = "awaiting_partner"
	AwaitingCompliment                    State = "awaiting_compliment"
	AwaitingComplimentFrequency           State = "awaiting_compliment_frequency"
	AwaitingTitleImportantDate            State = "awaiting_title_important_date"
	AwaitingDateImportantDate             State = "awaiting_date_important_date"
	AwaitingPartnerImportantDate          State = "awaiting_partner_important_date"
	AwaitingNotifyBeforeImportantDate     State = "awaiting_notify_before_important_date"
	AwaitingEditTitleImportantDate        State = "awaiting_edit_title_important_date"
	AwaitingEditDateImportantDate         State = "awaiting_edit_date_important_date"
	AwaitingEditPartnerImportantDate      State = "awaiting_edit_partner_important_date"
	AwaitingEditNotifyBeforeImportantDate State = "awaiting_edit_notify_before_important_date"
	AwaitingEditIsActiveImportantDate     State = "awaiting_edit_is_active_important_date"
)
