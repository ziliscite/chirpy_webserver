package main

import (
	"chirpy/database"
	"chirpy/handlers"
	"chirpy/helpers"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	// by default, godotenv will look for a file named .env in the current directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	JwtSecret := os.Getenv("JWT_SECRET")

	logger := helpers.NewLogger()
	mux := http.NewServeMux()

	db, err := database.NewDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
	}

	config := handlers.ApiConfig{
		FileServerHits: 0,
		DB:             db,
		JWTSecret:      JwtSecret,
	}

	mux.Handle("/app", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.StripPrefix("/app", http.FileServer(http.Dir("./"))))))
	mux.Handle("/assets", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.FileServer(http.Dir("./")))))
	mux.Handle("GET /api/health", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("OK"))
		if err != nil {
			return
		}
	}))))

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		logger.MiddlewareLogger(http.HandlerFunc(config.MetricsHandler)).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /api/reset", config.ResetMetrics)

	mux.Handle("POST /api/chirps", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.PostChirpsHandler))))
	mux.Handle("GET /api/chirps", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.GetChirpsHandler))))
	mux.Handle("GET /api/chirps/{id}", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.GetChirpHandler))))
	mux.Handle("DELETE /api/chirps/{id}", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.DeleteChirpsHandler))))

	mux.Handle("POST /api/users", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.RegisterUsersHandler))))
	mux.Handle("PUT /api/users", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.UpdateUsersHandler))))

	mux.Handle("POST /api/login", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.LoginHandler))))

	mux.Handle("POST /api/refresh", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.RefreshHandler))))
	mux.Handle("POST /api/revoke", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.RevokeTokenHandler))))

	mux.Handle("POST /api/polka/webhooks", config.MiddlewareMetricsInc(logger.MiddlewareLogger(http.HandlerFunc(config.PolkaHandler))))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		return
	}
}
