package services

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"invoice-generator/internal/auth"
	"invoice-generator/internal/constants"
	"invoice-generator/internal/database"
	"invoice-generator/internal/models"
)

// UserService handles user-related business logic
type UserService struct {
	jwtSecret []byte
}

// NewUserService creates a new user service
func NewUserService(jwtSecret []byte) *UserService {
	return &UserService{
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user
func (s *UserService) Register(user *models.User) error {
	// Validate email format
	if !strings.Contains(user.Email, "@") {
		return errors.New("invalid email format")
	}

	// Validate password length
	if len(user.Password) < constants.MinPasswordLength {
		return errors.New("password must be at least 6 characters")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Create user
	if err := database.GetDB().Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("email already exists")
		}
		return errors.New("failed to create user")
	}

	// Clear password before returning
	user.Password = ""
	return nil
}

// Login authenticates a user and returns a token
func (s *UserService) Login(loginData *models.LoginRequest) (string, *models.User, error) {
	var user models.User
	if err := database.GetDB().Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, user.IsAdmin, s.jwtSecret)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	// Clear password before returning
	user.Password = ""
	return token, &user, nil
}

// GetUsers returns all users (with access control)
func (s *UserService) GetUsers(isAdmin bool) ([]models.User, error) {
	var users []models.User
	query := database.GetDB().Select("id, email, name, company_name, gstin, address, city, state, pincode, phone, is_admin, created_at, updated_at")

	// If not admin, only return basic info
	if !isAdmin {
		query = query.Select("id, name, company_name, email")
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetProfile returns user profile by ID
func (s *UserService) GetProfile(userID uint) (*models.User, error) {
	var user models.User
	if err := database.GetDB().Select("id, email, name, company_name, gstin, address, city, state, pincode, phone, is_admin, created_at, updated_at").
		First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(userID uint, updateData *models.User) error {
	// Don't allow updating certain fields
	updateData.ID = 0
	updateData.Email = ""
	updateData.Password = ""
	updateData.IsAdmin = false

	if err := database.GetDB().Model(&models.User{}).Where("id = ?", userID).Updates(updateData).Error; err != nil {
		return errors.New("failed to update profile")
	}

	return nil
}
