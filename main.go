package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"gopkg.in/Clever/kayvee-go.v2/logger"
)

var kvlog = logger.New("graphviz-service")

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal("Missing env var: " + key)
	}
	return value
}

func main() {
	port := requiredEnv("PORT")

	startTime := time.Now()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		kvlog.Info("health-check")
		if time.Since(startTime) > 2*time.Minute {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(200)
	})

	http.HandleFunc("/dot", dotHandler)
	http.ListenAndServe(":"+port, nil)
}

func errResponse(w http.ResponseWriter, status int, body string, stats map[string]interface{}) {
	w.WriteHeader(status)
	w.Write([]byte(body))

	kvlog.ErrorD(body, stats)
}

func dotHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	stats := map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.String(),
		"remote": r.RemoteAddr,
		"start":  start.Format(time.RFC3339),
	}
	kvlog.InfoD("request-started", stats)

	if r.Method != "POST" {
		errResponse(w, 405, "Unknown method", stats)
		return
	}

	vals := r.URL.Query()["format"]

	format := "png"
	if len(vals) == 1 {
		format = vals[0]
	} else if len(vals) > 1 {
		errResponse(w, 400, "More than one format specified", stats)
		return
	}

	switch format {
	case "svg":
	case "png":
	case "pdf":
	case "plain":
	default:
		errResponse(w, 400, "Unkonwn format type", stats)
		return
	}

	dot := exec.Command("dot", "-T"+format)
	dot.Stdin = r.Body
	dot.Stdout = w
	dot.Stderr = os.Stderr

	err := dot.Run()
	if err != nil {
		errResponse(w, 500, err.Error(), stats)
		return
	}

	stats["duration-ms"] = time.Since(start) / time.Millisecond
	kvlog.InfoD("request-successful", stats)
}
