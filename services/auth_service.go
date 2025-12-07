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

	// 3. Generate verification token
	verificationToken, err := utils.GenerateNumericCode(6)
	if err != nil {
		logrus.Error("Error generating verification token: ", err)
		return nil, err
	}

	// 4. Build new user object
	newUser := models.User{
		ID:                  primitive.NewObjectID(),
		Name:                input.Name,
		Email:               input.Email,
		Password:            hashedPassword,
		Phone:               input.Phone,
		Role:                input.Role,
		AccountType:         input.AccountType,
		Status:              "active",
		Verified:            false,
		VerificationToken:   verificationToken,
		VerificationExpires: time.Now().Add(24 * time.Hour),
		ProfilePhoto:        input.ProfilePhoto,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Optional: Add role-specific sub-structs

	// switch input.Role {
	// case utils.ROLE_WORKER:
	// 	newUser.WorkerData = input.WorkerData
	// case utils.ROLE_CAR_OWNER:
	// 	newUser.OwnerData = input.OwnerData
	// }

	logrus.Info("Reached RegistrationUser Service")
	// 5. Insert into DB
	_, err = database.UserCollection.InsertOne(ctx, newUser)
	if err != nil {
		logrus.Error("Error inserting user: ", err)
		return nil, err
	}

	// 6. Send verification email (async)
	go func() {
		logrus.Infof("📧 [Email] Attempting to send verification email to %s", newUser.Email)
		if err := utils.SendVerificationEmail(newUser.Email, newUser.Name, verificationToken); err != nil {
			logrus.Errorf("❌ [Email] Failed to send verification email: %v", err)
		} else {
			logrus.Infof("✅ [Email] Verification email sent successfully to %s", newUser.Email)
		}
	}()

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

func (as *AuthService) VerifyEmail(email, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Find user
	user, err := as.userRepository.FindUserByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// 2. Check if already verified
	if user.Verified {
		return errors.New("email already verified")
	}

	// 3. Check token
	if user.VerificationToken != token {
		return errors.New("invalid verification code")
	}

	// 4. Check expiration
	if time.Now().After(user.VerificationExpires) {
		return errors.New("verification code expired")
	}

	// 5. Update user status
	update := primitive.M{
		"$set": primitive.M{
			"verified":           true,
			"verification_token": "", // Clear token
		},
	}

	_, err = database.UserCollection.UpdateOne(ctx, primitive.M{"_id": user.ID}, update)
	return err
}

// ResendVerificationEmail generates a new verification code and resends the email
func (as *AuthService) ResendVerificationEmail(email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Find user
	user, err := as.userRepository.FindUserByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// 2. Check if already verified
	if user.Verified {
		return errors.New("email already verified")
	}

	// 3. Generate new verification token
	verificationToken, err := utils.GenerateNumericCode(6)
	if err != nil {
		logrus.Error("Error generating verification token: ", err)
		return err
	}

	// 4. Update user with new token and expiration
	update := primitive.M{
		"$set": primitive.M{
			"verification_token":   verificationToken,
			"verification_expires": time.Now().Add(24 * time.Hour),
			"updated_at":           time.Now(),
		},
	}

	_, err = database.UserCollection.UpdateOne(ctx, primitive.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}

	// 5. Send verification email (async)
	go func() {
		if err := utils.SendVerificationEmail(user.Email, user.Name, verificationToken); err != nil {
			logrus.Error("Failed to resend verification email: ", err)
		}
	}()

	return nil
}
