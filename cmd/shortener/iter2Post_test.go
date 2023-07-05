package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostHandler(t *testing.T) {
	m = make(map[string]string)
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
				code:        201,
				response:    `http://localhost:8080/pqKqEKDz`,
				contentType: "text/plain",
			},
		},
		{
			name:   "Url encoder test#2",
			method: http.MethodGet,
			want: want{
				code:        400,
				response:    ``,
				contentType: "",
			},
		},
		{
			name:   "Url encoder test#3",
			body:   strings.NewReader("https://github.com/NuclearGhandi/http_project"),
			method: http.MethodPost,
			want: want{
				code:        201,
				response:    `http://localhost:8080/WfPgrmli`,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fmt.Println(test.body)
			request := httptest.NewRequest(test.method, "/status", test.body)

			// создаём новый Recorder
			w := httptest.NewRecorder()
			mainPage(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, test.want.code)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, string(resBody), test.want.response)
			assert.Equal(t, res.Header.Get("Content-Type"), test.want.contentType)
		})
	}
}
