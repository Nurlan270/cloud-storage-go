package hasher

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

// Default argon2id hashing parameters.
var params = &argon2id.Params{
	Memory:      32 * 1024, // 32 MiB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

//	Semaphore capacity = 2
//
// Max. 2 concurrent requests can hash password at the same time
// which will lead in 64 MiB memory usage; It's used to avoid memory overflow.
var sem = make(chan struct{}, 2)

func CreateHash(password string) (string, error) {
	sem <- struct{}{}
	defer func() { <-sem }()

	hash, err := argon2id.CreateHash(password, params)
	if err != nil {
		return "", fmt.Errorf("argon2id: failed to create hash: %w", err)
	}

	return hash, nil
}

func VerifyPassword(password, hash string) (bool, error) {
	sem <- struct{}{}
	defer func() { <-sem }()

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("argon2id: failed to compare password against hash: %w", err)
	}

	return match, nil
}
