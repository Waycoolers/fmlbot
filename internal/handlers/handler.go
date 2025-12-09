package handlers

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/storage"
	"github.com/Waycoolers/fmlbot/internal/ui"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	ui    *ui.MenuUI
	Store *storage.Storage
}

func New(ui *ui.MenuUI, store *storage.Storage) *Handler {
	return &Handler{ui: ui, Store: store}
}

func (h *Handler) ShowStartMenu(_ context.Context, chatID int64) {
	text := "Чтобы разбудить бота, зарегистрируйся по кнопке ниже"
	err := h.ui.StartMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить стартовое меню", err)
		return
	}
}

func (h *Handler) ShowMainMenu(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := "fmlbot приветствует тебя! 💖"
	err := h.ui.MainMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить главное меню", err)
		return
	}
}

func (h *Handler) Reply(chatID int64, text string) {
	err := h.ui.Client.SendMessage(chatID, text)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func (h *Handler) DoMidnightTasks(ctx context.Context) {
	err := h.Store.ClearComplimentsCount(ctx)
	if err != nil {
		log.Printf("Ошибка при очистке количества полученных комплиментов: %v", err)
	}

	err = h.Store.ClearComplimentTime(ctx)
	if err != nil {
		log.Printf("Ошибка при очистке времени последнего полученного комплимента: %v", err)
	}

	log.Print("Задачи выполнены")
}

func (h *Handler) ReplyUnknownCallback(_ context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	h.Reply(chatID, "Используй кнопки")
}

func (h *Handler) ReplyUnknownMessage(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	h.Reply(chatID, "Я не знаю такую команду")
}

func (h *Handler) HandleErr(chatID int64, msg string, err error) {
	h.Reply(chatID, "Произошла ошибка 😔")
	log.Printf("%s: %v", msg, err)
}
