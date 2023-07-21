package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

type Runtime struct {
	keytoURLMap map[string]string
	sugar       zap.SugaredLogger
}

type inputJSON struct {
	URL string `json:"url"`
}
type outputJSON struct {
	URL string `json:"result"`
}

var cfg Config
var rnt Runtime

func ServerInit() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	rnt.sugar = *logger.Sugar()
	rand.Seed(time.Now().UnixNano())
	serverAddressPointer := flag.String("a", ":8080", "Server Address")
	baseURLPointer := flag.String("b", "http://localhost:8080", "Base URL")
	flag.Parse()
	cfg.ServerAddress = *serverAddressPointer
	cfg.BaseURL = *baseURLPointer
	//fmt.Println(cfg.ServerAddress, "\n", cfg.BaseURL)
	err = env.Parse(&cfg)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "server init")
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func addURL(url string) string {
	key := randSeq(8)
	rnt.keytoURLMap[key] = url
	outURL := cfg.BaseURL + "/" + key
	return outURL
}
func handleGET(c *gin.Context) {
	key := c.Param("key")
	url, ok := rnt.keytoURLMap[key]
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
			c.String(http.StatusCreated, addURL(string(body)))
		}
	}
}
func handleAPIPOST(c *gin.Context) {
	fmt.Println("API")
	var inpt inputJSON
	var outpt outputJSON
	body, err := c.GetRawData()
	if err != nil {
		fmt.Println("API_body_get_err")
		serverErr(c)
	}
	if err = json.Unmarshal(body, &inpt); err != nil {
		fmt.Println("API_Unmsrshall_err")
		fmt.Println(err)
		serverErr(c)
	}
	outpt.URL = addURL(inpt.URL)
	resp, err := json.Marshal(outpt)
	if err != nil {
		fmt.Println("API_Marshall_err")
		serverErr(c)
	} else {
		c.Data(http.StatusCreated, "application/json", resp)
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Set("example", "12345")

		c.Next()
		method := c.Request.Method
		uri := c.Param("key")
		duration := time.Since(t)
		status := c.Writer.Status()
		size := c.Writer.Size()
		rnt.sugar.Infoln(
			"uri", uri,
			"method", method,
			"status", status,
			"duration", duration,
			"size", size,
		)
	}
}
func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(Logger())
	r.Use(gin.Recovery())

	r.GET("/:key", handleGET)
	r.POST("/", handlePOST)

	r.POST("/:key", serverErr)
	r.GET("/", serverErr)
	r.POST("/api/shorten", handleAPIPOST)
	r.GET("/api/:key", serverErr)
	return r
}

func serverErr(c *gin.Context) {
	c.AbortWithStatus(http.StatusBadRequest)
}
func main() {

	ServerInit()
	rnt.keytoURLMap = make(map[string]string)

	r := setupRouter()
	r.Run(cfg.ServerAddress)
}
