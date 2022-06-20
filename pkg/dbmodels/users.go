package dbmodels

import "gorm.io/gorm"

func GetUserByEmail(db *gorm.DB, email string) *User {
	user := &User{}
	err := db.Where("email = ?", email).First(user).Error
	if err != nil {
		return nil
	}

	return user
}
