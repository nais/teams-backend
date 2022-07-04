package dbmodels

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&ApiKey{},
		&AuditLog{},
		&Authorization{},
		&Correlation{},
		&ReconcileError{},
		&Role{},
		&RoleAuthorization{},
		&SystemState{},
		&System{},
		&TeamMetadata{},
		&Team{},
		&User{},
		&UserRole{},
		&UserTeam{},
	)
}
