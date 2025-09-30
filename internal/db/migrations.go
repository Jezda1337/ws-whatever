package db

import (
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
		if err := tx.AutoMigrate(); err != nil {
			return err
		}

		return nil
	})

	return migration.Migrate()
}
