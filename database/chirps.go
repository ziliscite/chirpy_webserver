package database

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
)

type Chirpy struct {
	Id     int    `json:"id"`
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

// CreateChirps creates a new chirp and saves it to disk
func (db *DB) CreateChirps(body string, userId string) (Chirpy, error) {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return Chirpy{}, fmt.Errorf("error loading database: %v", err)
	}

	newId := dbstruct.Id
	dbstruct.Id += 1

	chirpy := Chirpy{Id: newId, Body: body, UserId: userId}

	dbstruct.Chirps[newId] = chirpy
	if err = db.writeDB(dbstruct); err != nil {
		return Chirpy{}, fmt.Errorf("error writing chirps: %v", err)
	}

	return chirpy, nil
}

func (db *DB) DeleteChirpy(chirpyId int) error {
	mu := new(sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return fmt.Errorf("error loading database: %v", err)
	}

	// Authorization is in the damn handler, I don't give a fuck right now
	delete(dbstruct.Chirps, chirpyId)

	if err = db.writeDB(dbstruct); err != nil {
		return fmt.Errorf("error writing chirps: %v", err)
	}

	return nil
}

func sortChirps(method string, slice []Chirpy) {
	// Reminders that slices are passed by reference
	if method == "asc" || method == "" {
		slices.SortFunc(slice, func(a, b Chirpy) int {
			return cmp.Compare(a.Id, b.Id)
		})
	} else if method == "desc" {
		slices.SortFunc(slice, func(a, b Chirpy) int {
			return cmp.Compare(b.Id, a.Id)
		})
	}
}

// loadAndFilterChirps returns all chirps in the database and filer by option
func (db *DB) loadAndFilterChirps(method string, filterFunc func(Chirpy) bool) ([]Chirpy, error) {
	mu := new(sync.RWMutex)
	mu.RLock()
	defer mu.RUnlock()

	sliceChirps := make([]Chirpy, 0)

	dbstruct, err := db.loadDB()
	if err != nil {
		return sliceChirps, fmt.Errorf("error loading database: %v", err)
	}

	for _, v := range dbstruct.Chirps {
		if filterFunc(v) {
			sliceChirps = append(sliceChirps, v)
		}
	}

	sortChirps(method, sliceChirps)

	return sliceChirps, nil
}

func (db *DB) GetChirps(method string) ([]Chirpy, error) {
	return db.loadAndFilterChirps(method, func(_ Chirpy) bool {
		return true
	})
}

func (db *DB) GetChirpByAuthor(id string, method string) ([]Chirpy, error) {
	return db.loadAndFilterChirps(method, func(chirp Chirpy) bool {
		return chirp.UserId == id
	})
}

func (db *DB) GetChirp(id int) (Chirpy, error) {
	mu := new(sync.RWMutex)
	mu.RLock()
	defer mu.RUnlock()

	dbstruct, err := db.loadDB()
	if err != nil {
		return Chirpy{}, fmt.Errorf("error loading database: %v", err)
	}

	if _, ok := dbstruct.Chirps[id]; !ok {
		return Chirpy{}, fmt.Errorf("chirp with id %v not found", id)
	}

	return dbstruct.Chirps[id], nil
}
