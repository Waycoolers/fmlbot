package main

import (
	"log"

	"github.com/Waycoolers/fmlbot/internal/bot"
	"github.com/Waycoolers/fmlbot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	b.Run()
}
