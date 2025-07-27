// scripts/create_admin.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Email       string `gorm:"unique;not null"`
	Password    string `gorm:"not null"`
	Name        string `gorm:"not null"`
	CompanyName string
	GSTIN       string
	Address     string
	City        string
	State       string
	Pincode     string
	Phone       string
	IsAdmin     bool `gorm:"default:false"`
}

func main() {
	// Connect to database
	dsn := "host=localhost user=invoice_user password=invoice_password dbname=invoice_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Create Admin User ===")

	// Get user input
	fmt.Print("Enter email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Enter full name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter company name: ")
	companyName, _ := reader.ReadString('\n')
	companyName = strings.TrimSpace(companyName)

	fmt.Print("Enter GSTIN (optional): ")
	gstin, _ := reader.ReadString('\n')
	gstin = strings.TrimSpace(strings.ToUpper(gstin))

	fmt.Print("Enter phone: ")
	phone, _ := reader.ReadString('\n')
	phone = strings.TrimSpace(phone)

	// Get password securely
	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Failed to read password:", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create admin user
	adminUser := User{
		Email:       email,
		Password:    string(hashedPassword),
		Name:        name,
		CompanyName: companyName,
		GSTIN:       gstin,
		Phone:       phone,
		IsAdmin:     true,
	}

	// Check if user already exists
	var existingUser User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		fmt.Printf("User with email %s already exists!\n", email)
		return
	}

	// Create the user
	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Printf("Admin user created successfully!\n")
	fmt.Printf("Email: %s\n", adminUser.Email)
	fmt.Printf("Name: %s\n", adminUser.Name)
	fmt.Printf("Company: %s\n", adminUser.CompanyName)
}
