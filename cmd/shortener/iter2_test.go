package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		location    string
		contentType string
	}
	tests := []struct {
		name   string
		body   string
		URL    string
		Method http.methods
		want   want
	}{
		{
			name:   "Url encoder test#1",
			body:   "https://google.com",
			Method: http.MethodPost,
			want: want{
				code:        200,
				response:    `http://localhost:8080/pqKqEKDz`,
				contentType: "plain/text",
			},
		},
		{
			name:   "Url encoder test#2",
			body:   "https://google.com",
			Method: http.MethodPost,
			want: want{
				code:        200,
				response:    `http://localhost:8080/WfPgrmli`,
				contentType: "plain/text",
			},
		},
		{
			name:   "Url encoder test#3",
			body:   "https://github.com/NuclearGhandi/http_project",
			Method: http.MethodPost,
			want: want{
				code:        200,
				response:    `http://localhost:8080/WgRRVsgq`,
				contentType: "plain/text",
			},
		},
		{
			name:   "Url decoder&redirector test#1",
			URL:    `http://localhost:8080`,
			Method: http.MethodGet,
			want: want{
				code: 400,
			},
		},
		{
			name:   "Url decoder&redirector test#2",
			URL:    "http://localhost:8080/WgRRTgJu",
			Method: http.MethodPost,
			want: want{
				code: 400,
			},
		},
		{
			name:   "Url decoder&redirector test#3",
			URL:    "http://localhost:8080/WfPgrmli",
			Method: http.MethodGet,
			want: want{
				code:     307,
				location: "https://google.com",
			},
		},
		{
			name:   "Url decoder&redirector test#4",
			URL:    "http://localhost:8080/WgRRVsgq",
			Method: http.MethodGet,
			want: want{
				code:     307,
				location: "https://github.com/NuclearGhandi/http_project",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(Method, "/status", body)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			StatusHandler(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, test.want.code)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.JSONEq(t, string(resBody), test.want.response)
			assert.Equal(t, res.Header.Get("Content-Type"), test.want.contentType)
		})
	}
}
