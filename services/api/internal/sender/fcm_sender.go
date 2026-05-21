package sender

import (
	"context"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"google.golang.org/api/option"
)

type FCMSender struct {
	client *messaging.Client
	repo   domain.FCMRepo
}

func NewFCMSender(ctx context.Context, repo domain.FCMRepo, credentialsFile string) (domain.Sender, error) {
	opt := option.WithAuthCredentialsFile(
		option.ServiceAccount,
		credentialsFile,
	)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMSender{
		client: client,
		repo:   repo,
	}, nil
}

func (s *FCMSender) SendMessage(ctx context.Context, update domain.MessageRequest) error {
	token, err := s.repo.GetFCMToken(ctx, update.UserID)
	if err != nil {
		return err
	}

	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: "Новое сообщение",
			Body:  update.Text,
		},
	}

	_, err = s.client.Send(ctx, msg)
	return err
}

func (s *FCMSender) SendImportantDatesNotification(ctx context.Context, update domain.ImportantDateMessage) error {
	for _, userID := range update.UserIDs {
		token, err := s.repo.GetFCMToken(ctx, userID)
		if err != nil {
			continue
		}

		msg := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: "Важная дата",
				Body:  update.Message,
			},
		}

		_, err = s.client.Send(ctx, msg)
		if err != nil {
			continue
		}
	}

	return nil
}
