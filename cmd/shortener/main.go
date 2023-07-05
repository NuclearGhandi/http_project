package main

import (
	//	"fmt"
	"io"
	"math/rand"
	"net/http"
	// "time"
)

const host = "http://localhost:8080"

var m map[string]string

func init() {
	rand.Seed(100)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func postHandler(url string, m map[string]string) string {
	key := "/" + randSeq(8)
	m[key] = url
	outURL := host + key
	return outURL
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			serverErr(w)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, postHandler(string(body), m))
		}
	} else if r.Method == http.MethodGet {
		key := r.URL.Path
		url, ok := m[key]
		if ok {
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else {
			serverErr(w)
		}
	} else {
		serverErr(w)
	}
}

func redirectPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		key := r.URL.Path
		url, ok := m[key]
		if ok {
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else {
			serverErr(w)
		}
	} else {
		serverErr(w)
	}
}

func serverErr(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "text/html")
}
func main() {
	m = make(map[string]string)
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainPage)
	//	mux.HandleFunc(`/`, redirectPage)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
