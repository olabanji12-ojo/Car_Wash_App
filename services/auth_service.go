package services

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/olabanji12-ojo/CarWashApp/database"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/repositories"
	"github.com/olabanji12-ojo/CarWashApp/utils"
	// "fmt"
)

// proposed flow
// CAR OWNER
// Navigates to url, clicks on sign up and fills basic form
// proceeds to dashboard and can use the app, or maybe complete basic post onboarding like updating profile picture and perhaps funding wallet because we will add payment to it
// Navigates to url, clicks on sign up and fills basic form
// proceeds to post onboarding where he is presented another form to update business information
// then a virtual account is created for the business

type AuthService struct {
	userRepository repositories.UserRepository
}

func NewAuthService(userRepository repositories.UserRepository) *AuthService {
	return &AuthService{userRepository: userRepository}
}

func (as *AuthService) RegisterUser(input models.User) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Check if user with email already exists
	existing, _ := as.userRepository.FindUserByEmail(input.Email)
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	// 2. Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		logrus.Error("Error hashing password: ", err)
		return nil, err
	}

	// 3. Build new user object
	newUser := models.User{
		ID:           primitive.NewObjectID(),
		Name:         input.Name,
		Email:        input.Email,
		Password:     hashedPassword,
		Phone:        input.Phone,
		Role:         input.Role,
		AccountType:  input.AccountType,
		Status:       "active",
		Verified:     false,
		ProfilePhoto: input.ProfilePhoto,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Optional: Add role-specific sub-structs

	// switch input.Role {
	// case utils.ROLE_WORKER:
	// 	newUser.WorkerData = input.WorkerData
	// case utils.ROLE_CAR_OWNER:
	// 	newUser.OwnerData = input.OwnerData
	// }

	logrus.Info("Reached RegistrationUser Service")
	// 4. Insert into DB
	_, err = database.UserCollection.InsertOne(ctx, newUser)
	if err != nil {
		logrus.Error("Error inserting user: ", err)
		return nil, err
	}

	newUser.Password = ""
	return &newUser, nil
}

func (as *AuthService) LoginUser(email, password string) (string, *models.User, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Find user by email
	user, err := as.userRepository.FindUserByEmail(email)
	if err != nil {
		return "", nil, errors.New("user not found")
	}

	// 2. Check password
	if err := utils.CheckPasswordHash(password, user.Password); err != nil {
		return "", nil, errors.New("invalid password")
	}

	// 3. Generate JWT token
	token, err := utils.GenerateToken(user.ID.Hex(), user.Email, user.Role, user.AccountType)
	if err != nil {
		logrus.Error("Error generating token: ", err)
		return "", nil, err
	}

	// fmt.Println(user)

	return token, user, nil

}
