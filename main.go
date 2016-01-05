package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
)

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal("Missing env var: " + key)
	}
	return value
}

func main() {
	port := requiredEnv("PORT")

	http.HandleFunc("/dot", dotHandler)
	http.ListenAndServe(":"+port, nil)
}

func response(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	w.Write([]byte(body))
}

func dotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		response(w, 405, "Unknown method")
		return
	}

	vals := r.URL.Query()["format"]

	format := "png"
	if len(vals) == 1 {
		format = vals[0]
	} else if len(vals) > 1 {
		response(w, 400, "More than one format specified")
		return
	}

	switch format {
	case "svg":
	case "png":
	case "pdf":
	case "plain":
	default:
		response(w, 400, "Unkonwn format type")
		return
	}

	dot := exec.Command("dot", "-T"+format)
	dot.Stdin = r.Body
	dot.Stdout = w
	dot.Stderr = os.Stderr

	err := dot.Run()
	if err != nil {
		response(w, 500, err.Error())
		return
	}
}
