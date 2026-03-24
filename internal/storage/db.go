package storage

import (
	"fmt"
	"log"

	"github.com/Waycoolers/fmlbot/internal/config"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db             *sqlx.DB
	Compliments    *complimentsRepo
	ImportantDates *importantDatesRepo
	Scheduler      *schedulerRepo
	UserConfig     *userConfigRepo
	Users          *usersRepo
}

func New(cfg *config.DatabaseConfig) (*Storage, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	if er := db.Ping(); er != nil {
		log.Fatalf("Ошибка пинга в БД: %v", er)
	}

	log.Println("БД успешно подключена")

	compliments := complimentsRepo{
		db: db,
	}

	importantDates := importantDatesRepo{
		db: db,
	}

	scheduler := schedulerRepo{
		db: db,
	}

	userConfig := userConfigRepo{
		db: db,
	}

	users := usersRepo{
		db: db,
	}

	return &Storage{db: db,
		Compliments:    &compliments,
		ImportantDates: &importantDates,
		Scheduler:      &scheduler,
		UserConfig:     &userConfig,
		Users:          &users,
	}, nil
}

func (s *Storage) Close() {
	err := s.db.Close()
	if err != nil {
		log.Printf("Ошибка при закрытии подключения к бд: %v", err)
	}
}
