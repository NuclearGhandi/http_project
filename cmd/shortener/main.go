package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var m map[string]string

type сonfiguration struct {
	RunPort string
	host    string
}

var config сonfiguration

func ServerInit() {
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

func addURL(url string, m map[string]string) string {
	key := randSeq(8)
	m[key] = url
	outURL := config.host + "/" + key
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
	ServerInit()
	runPortPoint := flag.String("a", "8080", "RunPort")
	returnPortPoint := flag.String("b", "8080", "ReturnPort")
	flag.Parse()
	config.RunPort = *runPortPoint
	config.host = *returnPortPoint
	fmt.Println(*runPortPoint, "\n", *returnPortPoint)
	m = make(map[string]string)

	r := setupRouter()
	r.Run(config.RunPort)
}
