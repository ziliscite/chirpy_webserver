package database

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

type User struct {
	Id          string `json:"id"`
	Email       string `json:"email"`
	Password    []byte `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type RefreshToken struct {
	UserId   string    `json:"user_id"`
	Token    string    `json:"refresh_token"`
	ExpireAt time.Time `json:"expire_time"`
}

// CreateUsers creates a new user and saves it to disk
func (db *DB) CreateUsers(email string, password []byte) (User, int, error) {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, http.StatusInternalServerError, fmt.Errorf("error loading database: %v", err)
	}

	if _, ok := db.isEmailExist(dbstruct, email); ok {
		return User{}, http.StatusBadRequest, fmt.Errorf("email is already used")
	}

	newId, err := uuid.NewRandom()
	if err != nil {
		return User{}, http.StatusInternalServerError, fmt.Errorf("error creating new ID: %v", err)
	}

	id := newId.String()
	user := User{Id: id, Email: email, Password: password, IsChirpyRed: false}
	dbstruct.Users[id] = user

	err = db.writeDB(dbstruct)
	if err != nil {
		return user, http.StatusInternalServerError, fmt.Errorf("error writing user: %v", err)
	}

	return user, http.StatusCreated, nil
}

// GetUser returns a valid user by email address
func (db *DB) GetUser(email string) (User, int, error) {
	mu := new(sync.RWMutex)
	mu.RLock()
	defer mu.RUnlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, http.StatusInternalServerError, fmt.Errorf("error loading database: %v", err)
	}

	user, ok := db.isEmailExist(dbstruct, email)
	if !ok {
		return User{}, http.StatusUnauthorized, fmt.Errorf("email or password is invalid")
	}

	return user, http.StatusOK, nil
}

// UpdateUser returns a valid updated user or http.code and error message if the update failed
func (db *DB) UpdateUser(id string, newEmail string, newPassword []byte) (User, int, error) {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, http.StatusInternalServerError, fmt.Errorf("error loading database: %v", err)
	}

	// Aight, busted, bla bla, no auth in the db. Whatever man, I aint doing it
	if _, ok := db.isEmailExist(dbstruct, newEmail); ok {
		return User{}, http.StatusBadRequest, fmt.Errorf("email is already used")
	}

	user := User{Id: id, Email: newEmail, Password: newPassword}
	dbstruct.Users[id] = user

	err = db.writeDB(dbstruct)
	if err != nil {
		return User{}, http.StatusInternalServerError, fmt.Errorf("error writing user: %v", err)
	}

	return user, http.StatusOK, nil
}

func (db *DB) UpgradeUser(id string) (User, int, error) {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, http.StatusInternalServerError, fmt.Errorf("error loading database: %v", err)
	}

	user, ok := dbstruct.Users[id]
	if !ok {
		return User{}, http.StatusNotFound, fmt.Errorf("user not found")
	}

	user.IsChirpyRed = true
	dbstruct.Users[id] = user

	err = db.writeDB(dbstruct)
	if err != nil {
		return User{}, http.StatusInternalServerError, fmt.Errorf("error writing user: %v", err)
	}

	return user, http.StatusOK, nil
}

// isEmailExist returns user and true if the given email exists in db
func (db *DB) isEmailExist(dbstruct DBStruct, email string) (User, bool) {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	users := dbstruct.Users
	for _, v := range users {
		if v.Email != email {
			continue
		}

		return v, true
	}

	return User{}, false
}
