package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TServerInit() {
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	//	log.Println("test_init")
}
func TestPostRoute(t *testing.T) {
	rnt.keytoURLMap = make(map[string]string)
	type want struct {
		code        int
		response    string
		location    string
		contentType string
	}
	tests := []struct {
		name   string
		body   io.Reader
		URL    string
		method string
		want   want
	}{
		{
			name:   "Url encoder test#1",
			body:   strings.NewReader("https://google.com"),
			method: http.MethodPost,
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080`,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "Url encoder test#2",
			method: http.MethodGet,
			want: want{
				code:        http.StatusBadRequest,
				response:    ``,
				contentType: "",
			},
		},
		{
			name:   "Url encoder test#3",
			body:   strings.NewReader("https://github.com/NuclearGhandi/http_project"),
			method: http.MethodPost,
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080`,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	TServerInit()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := setupRouter()
			w := httptest.NewRecorder()
			req, err := http.NewRequest(test.method, "/", test.body)
			if err != nil {
				log.Fatal(err)
			}
			router.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if res.StatusCode == http.StatusCreated {
				fmt.Println(w.Body.String())
				assert.Equal(t, test.want.response, w.Body.String()[:21])
			}
		})
	}
}
