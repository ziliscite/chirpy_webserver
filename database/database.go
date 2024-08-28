package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

// ChirpyCounter To generate the correct chirpyId
// Will be incremented when a new chirpy is created
type ChirpyCounter struct {
	Id int `json:"id"`
}

type DBStruct struct {
	Chirps map[int]Chirpy          `json:"chirps"`
	Users  map[string]User         `json:"users"`
	Tokens map[string]RefreshToken `json:"refresh_tokens"`
	ChirpyCounter
}

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	db := DB{
		path: "database.json",
		mu:   &sync.RWMutex{},
	}

	// Check if the JSON file exists; otherwise create a new JSON file
	err := db.ensureDB()
	if err != nil {
		return nil, fmt.Errorf("error ensuring database exists: %s", err)
	}

	return &db, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		_, err = os.Create(db.path)
		if err != nil {
			return fmt.Errorf("error creating database file: %v", err)
		}

		dbstruct := DBStruct{
			Chirps:        map[int]Chirpy{},
			Users:         map[string]User{},
			Tokens:        map[string]RefreshToken{},
			ChirpyCounter: ChirpyCounter{Id: 1},
		}

		err = db.writeDB(dbstruct)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbstruct DBStruct) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbstruct)
	if err != nil {
		return fmt.Errorf("error marshalling json: %v", err)
	}

	// If the file does not exist, WriteFile creates it with permissions perm.
	err = os.WriteFile(db.path, dat, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStruct, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbstruct := DBStruct{
		Chirps:        map[int]Chirpy{},
		Users:         map[string]User{},
		Tokens:        map[string]RefreshToken{},
		ChirpyCounter: ChirpyCounter{Id: 1},
	}

	dat, err := os.ReadFile(db.path)
	if err != nil {
		return dbstruct, fmt.Errorf("error reading file: %v", err)
	}

	err = json.Unmarshal(dat, &dbstruct)
	if err != nil {
		return dbstruct, fmt.Errorf("error unmarshalling json: %v", err)
	}

	return dbstruct, nil
}
