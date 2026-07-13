package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// CreateUser inserts a new user record into the database.
func CreateUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

// GetUserByEmail finds a user by their email address.
func GetUserByEmail(db *gorm.DB, email string) (*model.User, error) {
	var user model.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user by their primary key ID.
func GetUserByID(db *gorm.DB, id string) (*model.User, error) {
	var user model.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser saves changes to an existing user record.
func UpdateUser(db *gorm.DB, user *model.User) error {
	return db.Save(user).Error
}

// EmailExists checks whether a user with the given email already exists.
func EmailExists(db *gorm.DB, email string) (bool, error) {
	var count int64
	if err := db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UsernameExists checks whether a user with the given username already exists.
func UsernameExists(db *gorm.DB, username string) (bool, error) {
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
