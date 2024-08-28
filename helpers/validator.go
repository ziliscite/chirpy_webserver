package helpers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"unicode"
)

func RequestBodyValidator(r *http.Request, requestBody any) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)

	if err != nil {
		log.Printf("Invalid request body: %s", err)
		return errors.New("invalid request body")
	}

	return nil
}

func IsValidEmail(email string) bool {
	// Basic regular expression for email validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func ValidatePassword(password string) error {
	var hasUpper, hasLower, hasNumber bool
	if len(password) < 8 {
		return errors.New("password is too short, it must contain at least 8 characters")
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain uppercase letters")
	}
	if !hasLower {
		return errors.New("password must contain lowercase letters")
	}
	if !hasNumber {
		return errors.New("password must contain numbers")
	}

	return nil
}
