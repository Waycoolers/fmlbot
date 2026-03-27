package handlers

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/redis_store"
	"github.com/Waycoolers/fmlbot/services/bot/internal/storage"
	"github.com/Waycoolers/fmlbot/services/bot/internal/ui"
)

type Handler struct {
	ui                      *ui.MenuUI
	Store                   *storage.Storage
	importantDateDrafts     *redis_store.ImportantDateDraftStore
	importantDateEditDrafts *redis_store.ImportantDateEditDraftStore
}

func New(ui *ui.MenuUI, store *storage.Storage, importantDateDrafts *redis_store.ImportantDateDraftStore, importantDateEditDrafts *redis_store.ImportantDateEditDraftStore) *Handler {
	return &Handler{ui: ui, Store: store, importantDateDrafts: importantDateDrafts, importantDateEditDrafts: importantDateEditDrafts}
}

func (h *Handler) ShowStartMenu(_ context.Context, chatID int64) {
	text := "✨ Чтобы разбудить бота, зарегистрируйся по кнопке ниже"
	err := h.ui.StartMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить стартовое меню", err)
		return
	}
}

func (h *Handler) ShowMainMenu(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	msgText := msg.Text
	text := "🌿 Выбери, что хочешь сделать"

	if msgText == string(domain.Register) {
		text = "bot приветствует тебя! 💖"
	}

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
	err := h.Store.Scheduler.ClearComplimentsCount(ctx)
	if err != nil {
		log.Printf("Ошибка при очистке количества полученных комплиментов: %v", err)
	}

	err = h.Store.Scheduler.ClearComplimentTokenBucket(ctx)
	if err != nil {
		log.Printf("Ошибка при очистке ведра для доступных комплиментов: %v", err)
	}

	log.Print("Задачи выполнены")
}

func (h *Handler) ReplyUnknownCallback(_ context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	h.Reply(chatID, "🤍 Лучше воспользуйся кнопками — так будет проще")
}

func (h *Handler) ReplyUnknownMessage(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	h.Reply(chatID, "🤔 Я пока не понимаю это сообщение\nПопробуй выбрать действие с кнопок ниже")
}

func (h *Handler) HandleErr(chatID int64, msg string, err error) {
	h.Reply(chatID, "😔 Что-то пошло не так\nЯ уже стараюсь всё исправить")
	log.Printf("%s: %v", msg, err)
}
