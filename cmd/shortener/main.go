package main

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

const host = "http://localhost:8080"

var m map[string]string

func init() {
	rand.Seed(125)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func addURL(url string, m map[string]string) string {
	key := randSeq(8)
	m[key] = url
	outURL := host + "/" + key
	return outURL
}

func handleGET(c *gin.Context) {
	key := c.Param("key")
	url, ok := m[key]
	if ok {
		c.Redirect(http.StatusTemporaryRedirect, url)
	} else {
		serverErr(c)
	}
}
func handlePOST(c *gin.Context) {
	if c.Param("key") != "" {
		serverErr(c)
	} else {
		body, err := c.GetRawData()
		if err != nil {
			serverErr(c)
		} else {
			c.String(http.StatusCreated, addURL(string(body), m))
		}
	}
}
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/:key", handleGET)
	r.POST("/", handlePOST)

	r.POST("/:key", serverErr)
	r.GET("/", serverErr)
	return r
}

func serverErr(c *gin.Context) {
	c.AbortWithStatus(http.StatusBadRequest)
}
func main() {
	m = make(map[string]string)
	r := setupRouter()
	r.Run(":8080")
}
