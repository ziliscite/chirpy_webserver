package helpers

import (
	"log"
	"net/http"
	"os"
)

type Logger struct {
	log *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		log: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile),
	}
}

func (logger *Logger) Println(v ...interface{}) {
	logger.log.Println(v...)
}

func (logger *Logger) Printf(format string, v ...interface{}) {
	logger.log.Printf(format, v...)
}

func (logger *Logger) Fatalf(format string, v ...interface{}) {
	logger.log.Fatalf(format, v...)
}

func (logger *Logger) Errorf(format string, v ...interface{}) {
	logger.log.Printf(format, v...)
}

func (logger *Logger) MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Println(r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
