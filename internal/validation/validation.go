package validation

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidName        = errors.New("نام فقط می‌تواند شامل حروف فارسی باشد و باید شامل نام و نام خانوادگی باشد ")
	ErrNameTooShort       = errors.New("نام باید حداقل ۵ حرف داشته باشد")
	ErrNameTooLong        = errors.New("نام نباید بیش از 50 حرف باشد")
	ErrInvalidPhone       = errors.New("شماره تلفن نامعتبر است و باید با ۰۹ شروع شود و ۱۱ رقم باشد")
	ErrInvalidCompanions  = errors.New("تعداد نفرات همراه باید بین ۰ تا ۵ باشد")
	ErrInvalidMajor       = errors.New("رشته تحصیلی فقط می‌تواند کامپیوتر یا برق باشد")
	ErrInvalidStudentID   = errors.New("شماره دانشجویی باید شامل ۱۰ تا ۱۵ رقم باشد")
	ErrStudentIDTooLong   = errors.New("شماره دانشجویی نباید بیش از 15 رقم باشد")
	ErrInvalidTransaction = errors.New("شماره رهگیری پرداخت باید شامل 5 تا 20 رقم باشد و فقط عدد باشد")
)

// ------------------- Name -------------------
func ValidateName(name string) error {
	name = strings.TrimSpace(name)

	length := len([]rune(name))
	if length < 5 {
		return ErrNameTooShort
	}
	if length > 50 {
		return ErrNameTooLong
	}
	if !strings.Contains(name, " ") {
		return ErrInvalidName
	}

	re := regexp.MustCompile(`^[\p{Arabic} ]+$`)
	if !re.MatchString(name) {
		return ErrInvalidName
	}

	return nil
}

// ------------------- Phone -------------------
func ValidatePhone(phone string) error {
	re := regexp.MustCompile(`^09\d{9}$`)
	if !re.MatchString(phone) {
		return ErrInvalidPhone
	}
	return nil
}

// ------------------- Companions -------------------
func ValidateCompanions(count int) error {
	if count < 0 || count > 5 {
		return ErrInvalidCompanions
	}
	return nil
}

// ------------------- Major -------------------
func ValidateMajor(major string) error {
	if major != "کامپیوتر" && major != "برق" {
		return ErrInvalidMajor
	}
	if len([]rune(major)) > 20 {
		return errors.New("طول رشته تحصیلی نباید بیش از 20 حرف باشد")
	}
	return nil
}

// ------------------- Student ID -------------------
func ValidateStudentID(id string) error {
	re := regexp.MustCompile(`^\d{10,15}$`)
	if !re.MatchString(id) {
		return ErrInvalidStudentID
	}
	if len(id) > 15 {
		return ErrStudentIDTooLong
	}
	return nil
}

// ------------------- Transaction -------------------
func ValidateTransaction(tx string) error {
	re := regexp.MustCompile(`^\d{5,20}$`)
	if !re.MatchString(tx) {
		return ErrInvalidTransaction
	}
	return nil
}

