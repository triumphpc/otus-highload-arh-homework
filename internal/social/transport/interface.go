package transport

type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}
