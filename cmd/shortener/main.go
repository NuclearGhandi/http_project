package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const host = "http://localhost:8080/"

var m map[string]string

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func encode(url string) string {
	key := randSeq(8)
	m[key] = url
	return key
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			serverErr(w)
		} else {
			fmt.Println(string(body))
			url := host + encode(string(body))
			fmt.Println(url)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, url)
		}
	} else if r.Method == http.MethodGet {
		key := r.URL.Path[1:]
		url, ok := m[key]
		if ok {
			fmt.Println(key)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else {
			serverErr(w)
		}
	}
}

// func redirectPage(w http.ResponseWriter, r *http.Request){
// key :=
// http.Redirect(w, r, m[key], http.StatusTemporaryRedirect)
// }
func serverErr(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "text/html")
}
func main() {
	m = make(map[string]string)
	//http.HandlerFunc('/',redirectPage)
	err := http.ListenAndServe(`:8080`, http.HandlerFunc(mainPage))
	if err != nil {
		panic(err)
	}
}
