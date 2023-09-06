package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostAPIRoute(t *testing.T) {
	rnt.keytoURLMap = make(map[string]string)
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		body   string
		err    error
		URL    string
		method string
		want   want
	}{
		{
			name: "Url encoder API test#1",

			body:   "https://google.com",
			URL:    "/api/shorten",
			method: http.MethodPost,
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080`,
				contentType: "application/json",
			},
		},
		{
			name:   "Url encoder API test#2",
			URL:    "/api/shorten",
			method: http.MethodGet,
			want: want{
				code:        http.StatusBadRequest,
				response:    ``,
				contentType: "",
			},
		},
		{
			name:   "Url encoder API test#3",
			body:   "https://vk.com",
			URL:    "/api/shorten",
			method: http.MethodPost,
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080`,
				contentType: "application/json",
			},
		},
	}
	TServerInit()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var inpt inputJSON
			var outpt outputJSON
			router := SetupRouter()
			w := httptest.NewRecorder()
			inpt.URL = test.body
			msg, err := json.Marshal(inpt)
			if err != nil {
				log.Fatal(err)
			}
			reader := bytes.NewReader(msg)
			req, err := http.NewRequest(test.method, test.URL, reader)
			if err != nil {
				log.Fatal(err)
			}
			router.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			if err != nil {
				log.Fatal(err)
			}
			if res.StatusCode == http.StatusCreated {
				if err = json.Unmarshal(w.Body.Bytes(), &outpt); err != nil {
					log.Fatal(err)
				}
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

				//fmt.Println(w.Body.String())
				//fmt.Println(outpt.URL)
				assert.Equal(t, test.want.response, outpt.URL[:21])
			}
			defer res.Body.Close()
		})
	}
}
