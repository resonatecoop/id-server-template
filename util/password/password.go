package password

import (
	"fmt"
	"github.com/apokalyptik/phpass"
	"golang.org/x/crypto/bcrypt"
	"os"
	"sync"
)

var (
	phpassVar   = phpass.New(phpass.NewConfig())
	phpassMutex = &sync.Mutex{}
)

// VerifyPassword compares password and the hashed password
// Fallback to phpass if bcrypt fails
func VerifyPassword(passwordHash, password string) error {
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		fmt.Fprintln(os.Stderr, "Not Bcrypt password")
		phpassMutex.Lock()
		ret := phpassVar.Check([]byte(password), []byte(passwordHash))
		phpassMutex.Unlock()
		// If passwords don't match return mismatch error
		// use bcrypt's error code for compat
		if !ret {
			fmt.Fprintln(os.Stderr, "No password match")
			return bcrypt.ErrMismatchedHashAndPassword
		}
		fmt.Fprintln(os.Stderr, "PHPassword matched")
		return nil
	}
	fmt.Fprintln(os.Stderr, "Bcrypt Password matched")
	return nil
}

// HashPassword creates a bcrypt password hash
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 3)
}

// HashWpPassword creates a phpass password hash
func HashWpPassword(password string) ([]byte, error) {
	phpassMutex.Lock()
	passwordHashWp, err := phpassVar.Hash([]byte(password))
	phpassMutex.Unlock()
	if err != nil {
		return nil, err
	}
	return passwordHashWp, nil
}
