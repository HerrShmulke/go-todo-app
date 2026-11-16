package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      uint32 = 19 * 1024
	iterations  uint32 = 2
	parallelism uint8  = 1

	saltLength = 16
	keyLength  = 32
)

var ErrPasswordMismatch = errors.New("password mismatch")

type Argon2Hasher struct{}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{}
}

func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, saltLength)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func (h *Argon2Hasher) Compare(encodedHash, password string) error {
	params, salt, expectedHash, err := decodeHash(encodedHash)
	if err != nil {
		return fmt.Errorf("failed to decode password hash: %w", err)
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(expectedHash, actualHash) != 1 {
		return ErrPasswordMismatch
	}

	return nil
}

type argon2Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

func decodeHash(encoded string) (argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")

	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return argon2Params{}, nil, nil, errors.New("invalid hash format")
	}

	version, err := strconv.Atoi(strings.TrimPrefix(parts[2], "v="))
	if err != nil || version != argon2.Version {
		return argon2Params{}, nil, nil, errors.New("unsupported argon2 version")
	}

	var params argon2Params

	for _, param := range strings.Split(parts[3], ",") {
		keyValue := strings.SplitN(param, "=", 2)

		if len(keyValue) != 2 {
			return argon2Params{}, nil, nil, errors.New("invalid parameters")
		}

		value, err := strconv.ParseUint(keyValue[1], 10, 32)
		if err != nil {
			return argon2Params{}, nil, nil, errors.New("invalid parameter value")
		}

		switch keyValue[0] {
		case "m":
			params.memory = uint32(value)
		case "t":
			params.iterations = uint32(value)
		case "p":
			if value > 255 {
				return argon2Params{}, nil, nil, errors.New("invalid parallelism")
			}
			params.parallelism = uint8(value)
		default:
			return argon2Params{}, nil, nil, errors.New("unknown parameter")
		}
	}

	// Prevent excessive resource consumption from malformed hashes.
	if params.memory < 8*uint32(params.parallelism) ||
		params.memory > 64*1024 ||
		params.iterations == 0 ||
		params.iterations > 10 ||
		params.parallelism == 0 ||
		params.parallelism > 8 {
		return argon2Params{}, nil, nil, errors.New("parameters out of bounds")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 64 {
		return argon2Params{}, nil, nil, errors.New("invalid salt")
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(hash) < 16 || len(hash) > 64 {
		return argon2Params{}, nil, nil, errors.New("invalid hash")
	}

	return params, salt, hash, nil
}
