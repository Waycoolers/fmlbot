package domain

type Command string

const (
	Start         Command = "/start"
	Main          Command = "🏠 На главную"
	Account       Command = "⚙️ Мой аккаунт"
	Register      Command = "✨ Начать"
	DeleteAccount Command = "🗑️ Удалить аккаунт"

	Partner       Command = "👤 Мой партнёр"
	AddPartner    Command = "➕ Добавить партнёра"
	DeletePartner Command = "➖ Удалить партнёра"

	Compliments             Command = "❤️ Комплименты"
	AddCompliment           Command = "💌 Добавить комплимент"
	DeleteCompliment        Command = "🗑️ Удалить комплимент"
	GetCompliments          Command = "📜 Все комплименты"
	ReceiveCompliment       Command = "✨ Получить комплимент"
	EditComplimentFrequency Command = "⏰ Лимит в день"

	ImportantDates      Command = "📅 Важные даты"
	AddImportantDate    Command = "➕ Добавить дату"
	GetImportantDates   Command = "📖 Мои даты"
	DeleteImportantDate Command = "🗑️ Удалить дату"
	EditImportantDate   Command = "✏️ Управление"
)

var Commands = []Command{
	Start,
	Main,
	Account,
	Register,
	DeleteAccount,
	Partner,
	AddPartner,
	DeletePartner,
	Compliments,
	AddCompliment,
	DeleteCompliment,
	GetCompliments,
	ReceiveCompliment,
	EditComplimentFrequency,
	ImportantDates,
	AddImportantDate,
	GetImportantDates,
	DeleteImportantDate,
	EditImportantDate,
}
