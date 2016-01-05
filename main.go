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

func dotHandler(w http.ResponseWriter, r *http.Request) {
	dot := exec.Command("dot", "-Tpng")
	dot.Stdin = r.Body
	dot.Stdout = w
	dot.Stderr = os.Stderr

	err := dot.Run()
	if err != nil {
		log.Fatal("ERROR: " + err.Error())
	}
}
