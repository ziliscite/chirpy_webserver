package handlers

import (
	"chirpy/database"
	"fmt"
	"net/http"
)

type ApiConfig struct {
	FileServerHits int
	DB             *database.DB
	JWTSecret      string
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	// so, we do not just `return next`, but to literally return something like in main, with the logger and all
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf(`
	<html>
		<title>admin</title>
		<body>
			<h1>Welcome, Chirpy Admin</h1>    
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`, cfg.FileServerHits)))
	if err != nil {
		return
	}
}

func (cfg *ApiConfig) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.FileServerHits = 0
	_, err := w.Write([]byte("Hits reset to 0"))
	if err != nil {
		return
	}
}
