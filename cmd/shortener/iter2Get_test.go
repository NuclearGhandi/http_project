package main

import (
	//	"fmt"
	//	"io"
	"net/http"
	"net/http/httptest"

	//	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetTestInit() map[string]string {
	m = map[string]string{
		"/WegaFuIk": "https://vk.com",
		"/llllllll": "https://vk.com",
		"/GGGGGGGG": "https://google.com",
	}
	return m
}
func TestGetHandler(t *testing.T) {
	m = GetTestInit()
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
			URL:    "/WegaFuIk",
			method: http.MethodGet,
			want: want{
				code:     307,
				location: "https://vk.com",
			},
		},
		{
			name:   "Url decoder test#2",
			URL:    "/llllllll",
			method: http.MethodGet,
			want: want{
				code:     307,
				location: "https://vk.com",
			},
		},
		{
			name:   "Url decoder test#3",
			URL:    "/GGGGGGGG",
			method: http.MethodGet,
			want: want{
				code:     307,
				location: "https://google.com",
			},
		},
		{
			name:   "Url decoder test#4",
			URL:    "/gfjdhyul",
			method: http.MethodGet,
			want: want{
				code: 400,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.URL, nil)

			// создаём новый Recorder
			w := httptest.NewRecorder()
			mainPage(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, test.want.code)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			if res.StatusCode == 307 {
				inUrl, err := res.Location()
				require.NoError(t, err)
				assert.Equal(t, inUrl.String(), test.want.location)
			}
		})
	}
}
