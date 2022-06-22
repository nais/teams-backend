package dbmodels

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&ApiKey{},
		&AuditLog{},
		&Correlation{},
		&ReconcileError{},
		&SystemState{},
		&System{},
		&SystemsTeams{},
		&TeamMetadata{},
		&Team{},
		&User{},
		&UsersTeams{},
	)
}
