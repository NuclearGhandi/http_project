package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	flagsRead       int
	flagsWrtie      int
}

type Runtime struct {
	keytoURLMap map[string]string
	sugar       zap.SugaredLogger
	fileLen     int
}

type fileJSON struct {
	UUID        int    `json:"uuid,string"`
	ShortURL    string `json:"short_url,string"`
	OriginalURL string `json:"original_url,string"`
}

type inputJSON struct {
	URL string `json:"url"`
}
type outputJSON struct {
	URL string `json:"result"`
}

var cfg Config
var rnt Runtime

func FileInit() {
	var file *os.File
	var err error
	file, err = os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
		fmt.Println("init_err")
	}
	file.Close()
}
func MapInit() {
	var file *os.File
	var scanner *bufio.Scanner
	var err error
	var buf fileJSON
	file, err = os.OpenFile(cfg.FileStoragePath, cfg.flagsRead, 0666)
	if err != nil {
		log.Fatal(err)
	}
	scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		data := scanner.Bytes()
		if err = json.Unmarshal(data, &buf); err != nil {
			fmt.Println(err)
		}
		rnt.fileLen = buf.UUID
		rnt.keytoURLMap[buf.ShortURL] = buf.OriginalURL
	}
	file.Close()
}

func FileWrite(shortURL string, originalURL string) {
	if cfg.FileStoragePath != "" {
		var file *os.File
		var outpt fileJSON
		outpt.OriginalURL = originalURL
		outpt.ShortURL = shortURL
		outpt.UUID = rnt.fileLen
		rnt.fileLen++
		data, err := json.Marshal(outpt)
		data = append(data, '\n')
		if err != nil {
			log.Fatal(err)
		}
		file, err = os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		_, err = file.Write(data)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
}

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
	FileStoragePathPointer := flag.String("f", "/tmp/short-url-db.json", "File Path")
	flag.Parse()
	cfg.ServerAddress = *serverAddressPointer
	cfg.BaseURL = *baseURLPointer
	cfg.FileStoragePath = *FileStoragePathPointer
	err = env.Parse(&cfg)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "server init")
	}
	cfg.flagsRead = os.O_RDONLY | os.O_CREATE
	cfg.flagsWrtie = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if cfg.FileStoragePath != "" {
		FileInit()
		MapInit()
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
	fmt.Println("map key", outURL, "value", url)
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
		header := c.Request.Header
		duration := time.Since(t)
		status := c.Writer.Status()
		size := c.Writer.Size()
		rnt.sugar.Infoln(
			"uri", uri,
			"method", method,
			"header", header,
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
	r.Use(Gzip(DefaultCompression))

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
