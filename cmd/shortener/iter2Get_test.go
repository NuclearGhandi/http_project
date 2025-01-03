package main

import (
	//	"fmt"
	//	"io"

	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRoute(t *testing.T) {
	var urls [3]string
	urls[0] = "https://vk.com"
	urls[1] = "https://vk.com"
	urls[2] = "https://google.com"
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name   string
		URL    string
		method string
		want   want
	}{
		{
			name:   "Url decoder test#1",
			URL:    "",
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://vk.com",
			},
		},
		{
			name:   "Url decoder test#2",
			URL:    "",
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://vk.com",
			},
		},
		{
			name:   "Url decoder test#3",
			URL:    "",
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://google.com",
			},
		},
		{
			name:   "Url decoder test#4",
			URL:    "/12012412",
			method: http.MethodGet,
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	TServerInit()
	router := SetupRouter()
	rnt.keytoURLMap = make(map[string]string)
	for i, URL := range urls {
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader(URL))
		if err != nil {
			log.Fatal(err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		tests[i].URL = w.Body.String()[21:]
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.method, test.URL, nil)
			router.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode == http.StatusTemporaryRedirect {
				inURL, err := res.Location()
				require.NoError(t, err)
				assert.Equal(t, inURL.String(), test.want.location)
			}
			defer res.Body.Close()
		})
	}
}

func TestGetRouteCompress(t *testing.T) {
	var urls [3]string
	urls[0] = "https://discord.com"
	urls[1] = "https://discord.com"
	urls[2] = "https://google.com"
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name   string
		URL    string
		method string
		want   want
	}{
		{
			name:   "Url decoder test#1",
			URL:    "",
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://discord.com",
			},
		},
		{
			name:   "Url decoder test#2",
			URL:    "",
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://discord.com",
			},
		},
		{
			name:   "Url decoder test#3",
			URL:    "",
			method: http.MethodGet,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://google.com",
			},
		},
		{
			name:   "Url decoder test#4",
			URL:    "/12012412",
			method: http.MethodGet,
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	TServerInit()
	router := SetupRouter()
	rnt.keytoURLMap = make(map[string]string)
	for i, URL := range urls {
		b, err := Compress([]byte(URL))
		if err != nil {
			log.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", bytes.NewReader(b))
		req.Header.Add("Content-Encoding", "gzip")
		if err != nil {
			log.Fatal(err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		tests[i].URL = w.Body.String()[21:]
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.method, test.URL, nil)
			//req.Header.Add("Accept-Encoding", "gzip")
			router.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode == http.StatusTemporaryRedirect {
				inURL, err := res.Location()
				require.NoError(t, err)
				assert.Equal(t, inURL.String(), test.want.location)
			}
			defer res.Body.Close()
		})
	}
}
