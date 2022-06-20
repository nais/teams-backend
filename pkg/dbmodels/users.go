package dbmodels

import (
	"gorm.io/gorm"
	"strings"
)

func GetUserByEmail(db *gorm.DB, email string) *User {
	user := &User{}
	err := db.Where("email = ?", strings.ToLower(email)).First(user).Error
	if err != nil {
		return nil
	}

	return user
}
