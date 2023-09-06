package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed decompress: %v", err)
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
func TServerInit() {
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.typeOfStorage = "map"
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	rnt.sugar = *logger.Sugar()
	//	log.Println("test_init")
}
func TestPostRoute(t *testing.T) {
	rnt.keytoURLMap = make(map[string]string)
	type want struct {
		code        int
		response    string
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
			router := SetupRouter()
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
			defer res.Body.Close()
		})
	}
}
