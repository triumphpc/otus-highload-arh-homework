package auth

import "golang.org/x/crypto/bcrypt"

// BcryptHasher реализует PasswordHasher с использованием bcrypt
type BcryptHasher struct {
	cost int // Стоимость хеширования (от 4 до 31)
}

// NewBcryptHasher создает новый экземпляр BcryptHasher
// Рекомендуемые значения cost:
// - 10 для разработки
// - 12-14 для production
func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// Hash создает bcrypt хеш пароля
func (h *BcryptHasher) Hash(password string) (string, error) {
	// Bcrypt имеет ограничение на длину пароля (72 байта)
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// Check сравнивает пароль с хешем
func (h *BcryptHasher) Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsHashed проверяет, соответствует ли строка формату bcrypt хеша
func (h *BcryptHasher) IsHashed(hash string) bool {
	// Bcrypt хеш начинается с $2a$, $2b$, $2x$ или $2y$
	return len(hash) >= 60 &&
		(hash[:4] == "$2a$" ||
			hash[:4] == "$2b$" ||
			hash[:4] == "$2x$" ||
			hash[:4] == "$2y$")
}
