package utils

import (
	"golang.org/x/crypto/bcrypt"
	"github.com/sirupsen/logrus"
)

// HashPassword takes a plain password and returns the hashed version
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Error("Error hashing password: ", err)
		return "", err
	}
	logrus.Info("Password hashed successfully")
	return string(hashed), nil
}


// CheckPasswordHash compares a plain password with the hashed one
func CheckPasswordHash(password, hashed string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		logrus.Warn("Password comparison failed")
		return err
	}
	logrus.Info("Password match confirmed")
	return nil
}



