package domain

type Command string

const (
	Start                   Command = "/start"
	Main                    Command = "↩️ Главная"
	Account                 Command = "⚙ Аккаунт"
	Partner                 Command = "👤 Партнёр"
	Compliments             Command = "❤️ Комплименты"
	ImportantDates          Command = "Важные даты"
	Register                Command = "Зарегистрироваться"
	DeleteAccount           Command = "Удалить аккаунт"
	AddPartner              Command = "Добавить партнёра"
	DeletePartner           Command = "Удалить партнёра"
	AddCompliment           Command = "Добавить комплимент"
	DeleteCompliment        Command = "Удалить комплимент"
	GetCompliments          Command = "Все комплименты"
	ReceiveCompliment       Command = "Получить комплимент"
	EditComplimentFrequency Command = "Лимит в день"
	AddImportantDate        Command = "Добавить важную дату"
	GetImportantDates       Command = "Мои важные даты"
	DeleteImportantDate     Command = "Удалить важную дату"
)
