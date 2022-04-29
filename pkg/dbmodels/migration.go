package dbmodels

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&ApiKey{},
		&AuditLog{},
		&Role{},
		&RoleBinding{},
		&Synchronization{},
		&System{},
		&TeamMetadata{},
		&Team{},
		&User{},
	)
}
