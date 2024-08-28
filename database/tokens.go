package database

import (
	"errors"
	"sync"
	"time"
)

func (db *DB) StoreRefreshToken(refreshToken RefreshToken) error {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return err
	}

	// If the user with that id already have a refresh token,
	// we revoke the previous refresh token to avoid duplicates
	for _, v := range dbstruct.Tokens {
		if v.UserId != refreshToken.UserId {
			continue
		}

		err = db.RevokeRefreshToken(v.Token)
		if err != nil {
			return err
		}

		break
	}

	// Why not user id as the key? Because refresh token (like, the string)
	// is the one that will be present in the request header
	dbstruct.Tokens[refreshToken.Token] = refreshToken

	err = db.writeDB(dbstruct)
	if err != nil {
		return err
	}

	return nil
}

// RevokeRefreshToken Takes refresh token string as a key to revoke a refresh token and return an error
func (db *DB) RevokeRefreshToken(refreshToken string) error {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return err
	}

	delete(dbstruct.Tokens, refreshToken)

	err = db.writeDB(dbstruct)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetRefreshToken(refreshToken string) (RefreshToken, error) {
	mu := new(sync.RWMutex)
	mu.RLock()
	defer mu.RUnlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return RefreshToken{}, err
	}

	token, ok := dbstruct.Tokens[refreshToken]
	if !ok {
		return RefreshToken{}, errors.New("refresh token not found")
	}

	// This is the code that actually returns the token (if not expire)
	if token.ExpireAt.After(time.Now()) {
		return token, nil
	}

	// Inversion, bro
	err = db.RevokeRefreshToken(refreshToken)
	if err != nil {
		return RefreshToken{}, err
	}

	return RefreshToken{}, errors.New("refresh token has expired and is revoked")
}
