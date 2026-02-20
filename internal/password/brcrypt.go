package password

import "golang.org/x/crypto/bcrypt"

var dummyHash = []byte("$2a$12$za9sI6H/KEm4ds8ZtVkPEecFr6XC.xKRJ5n0e7Maf.Lh1mtg0U9Qq")

func Hash(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func CompareDummy(plainText string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plainText))
}
