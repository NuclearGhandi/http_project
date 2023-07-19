package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

type Runtime struct {
	keyToUrlMap map[string]string
}

var cfg Config
var rnt Runtime

func ServerInit() {
	rand.Seed(time.Now().UnixNano())
	serverAddressPointer := flag.String("a", ":8080", "Server Address")
	baseURLPointer := flag.String("b", "http://localhost:8080", "Base URL")
	flag.Parse()
	cfg.ServerAddress = *serverAddressPointer
	cfg.BaseURL = *baseURLPointer
	fmt.Println(cfg.ServerAddress, "\n", cfg.BaseURL)
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(cfg)
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
	outURL := cfg.BaseURL + "/" + key
	return outURL
}

func handleGET(c *gin.Context) {
	key := c.Param("key")
	url, ok := rnt.keyToUrlMap[key]
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
			c.String(http.StatusCreated, addURL(string(body), rnt.keyToUrlMap))
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
	rnt.keyToUrlMap = make(map[string]string)

	r := setupRouter()
	r.Run(cfg.ServerAddress)
}
