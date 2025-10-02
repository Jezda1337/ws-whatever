package db

import (
	"ws-whatever/ws"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var migrations = []*gormigrate.Migration{
	// {
	// 	ID: "20250930225731_0_0_1__init_schema",
	// 	Migrate: func(db *gorm.DB) error {
	// 		return db.Migrator().AutoMigrate(&domain.Messages{})
	// 	},
	// 	Rollback: func(d *gorm.DB) error {
	//
	// 	},
	// },
}

func RunMigration(db *gorm.DB) error {
	migration := gormigrate.New(db, gormigrate.DefaultOptions, migrations)

	migration.InitSchema(func(tx *gorm.DB) error {
		if err := tx.Exec("CREATE TYPE room_type AS ENUM ('group', 'direct')").Error; err != nil {
			return err
		}

		if err := tx.AutoMigrate(
			&ws.Community{},
			&ws.User{},
			&ws.Room{},
			&ws.RoomParticipant{},
			&ws.Message{},
			&ws.MessageAttachment{},
			&ws.MessageReaction{},
			&ws.MessageRead{},
		); err != nil {
			return err
		}

		if err := tx.Create(&ws.Community{ID: 1}).Error; err != nil {
			return err
		}

		if err := tx.Create(&ws.User{ID: 1}).Error; err != nil {
			return err
		}

		return nil
	})

	return migration.Migrate()
}
