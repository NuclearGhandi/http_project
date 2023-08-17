package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"

	//	"fmt"
	//	"log"
	"database/sql"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type Config struct {
	typeOfStorage   string
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

type Runtime struct {
	keytoURLMap map[string]string
	sugar       zap.SugaredLogger
	fileLen     int
	db          *sql.DB
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

func DatabaseInit() {
	buf, err := sql.Open("postgres", cfg.DatabaseDSN)
	rnt.db = buf
	fmt.Println("DB Init")
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "databaseInit")
	}
	_, errr := rnt.db.Exec("CREATE TABLE IF NOT EXISTS shorted (seq VARCHAR(10) PRIMARY KEY, url VARCHAR(2084))")
	if errr != nil {
		rnt.sugar.Fatalw(errr.Error(), "event", "dbInit")
	}
}

func dbWriteURL(key string, url string) {
	_, err := rnt.db.Exec("INSERT INTO shorted ( seq, url) VALUES ($1, $2)", key, url)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "dbWrite")
	}
	fmt.Println(key, url)
	fmt.Println(dbReadURL(key))

}

func dbReadURL(key string) string {
	var url string
	row := rnt.db.QueryRow(
		"SELECT URL FROM shorted WHERE seq = $1", key)
	err := row.Scan(&url)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "dbRead")
	}
	return url
}

func FileDBTransfer() {
	var file *os.File
	var scanner *bufio.Scanner
	var err error
	var buf fileJSON
	file, err = os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "FileReadOpen")
	}
	scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		data := scanner.Bytes()
		if err = json.Unmarshal(data, &buf); err != nil {
			rnt.sugar.Fatalw(err.Error(), "event", "FileReadMarshalErr")
		}
		rnt.fileLen = buf.UUID
		dbWriteURL(buf.ShortURL, buf.OriginalURL)
	}
	file.Close()
}

func FileInit() {
	var file *os.File
	var err error
	file, err = os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "FileInit")
	}
	file.Close()
}
func MapInit() {
	var file *os.File
	var scanner *bufio.Scanner
	var err error
	var buf fileJSON
	file, err = os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "FileReadOpen")
	}
	scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		data := scanner.Bytes()
		if err = json.Unmarshal(data, &buf); err != nil {
			rnt.sugar.Fatalw(err.Error(), "event", "FileReadMarshalErr")
		}
		rnt.fileLen = buf.UUID
		rnt.keytoURLMap[buf.ShortURL] = buf.OriginalURL
	}
	file.Close()
}

func FileWrite(shortURL string, originalURL string) {
	var file *os.File
	var outpt fileJSON
	outpt.OriginalURL = originalURL
	outpt.ShortURL = shortURL
	outpt.UUID = rnt.fileLen
	rnt.fileLen++
	data, err := json.Marshal(outpt)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "FileMarshal")
	}
	data = append(data, '\n')
	file, err = os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "FileWriteOpen")
	}
	_, err = file.Write(data)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "FileWrite")
	}
	file.Close()
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
	DatabaseDSNPointer := flag.String("d", "", "Database DSN")
	flag.Parse()
	cfg.ServerAddress = *serverAddressPointer
	cfg.BaseURL = *baseURLPointer
	cfg.FileStoragePath = *FileStoragePathPointer
	cfg.DatabaseDSN = *DatabaseDSNPointer
	err = env.Parse(&cfg)
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "ServerInit")
	}
	if cfg.DatabaseDSN != "" {
		cfg.typeOfStorage = "db"
		DatabaseInit()
		if cfg.FileStoragePath != "" {
			FileInit()
			FileDBTransfer()
		}
	} else if cfg.FileStoragePath != "" {
		cfg.typeOfStorage = "file"
		FileInit()
		MapInit()
	} else {
		cfg.typeOfStorage = "map"
	}
	fmt.Println(cfg.DatabaseDSN)
	fmt.Println(cfg.typeOfStorage)
	fmt.Println(cfg.FileStoragePath)
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
	if cfg.typeOfStorage == "file" || cfg.typeOfStorage == "map" {
		rnt.keytoURLMap[key] = url
	}
	if cfg.typeOfStorage == "file" {
		FileWrite(key, url)
	}
	if cfg.typeOfStorage == "db" {
		fmt.Println("db write")
		dbWriteURL(key, url)
	}
	outURL := cfg.BaseURL + "/" + key
	return outURL
}
func handleGET(c *gin.Context) {
	key := c.Param("key")
	if cfg.typeOfStorage == "map" || cfg.typeOfStorage == "file" {
		url, ok := rnt.keytoURLMap[key]
		if ok {
			c.Redirect(http.StatusTemporaryRedirect, url)
		} else {
			serverErr(c)
		}
	} else if cfg.typeOfStorage == "db" {
		url := dbReadURL(key)
		if url != "" {
			c.Redirect(http.StatusTemporaryRedirect, url)
		} else {
			serverErr(c)
		}
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
			if cfg.typeOfStorage != "db" {
				c.String(http.StatusCreated, addURL(string(body)))
			}
		}
	}
}
func handleAPIPOST(c *gin.Context) {
	var inpt inputJSON
	var outpt outputJSON
	body, err := c.GetRawData()
	if err != nil {
		serverErr(c)
	}
	if err = json.Unmarshal(body, &inpt); err != nil {
		serverErr(c)
	}
	outpt.URL = addURL(inpt.URL)
	resp, err := json.Marshal(outpt)
	if err != nil {
		serverErr(c)
	} else {
		c.Data(http.StatusCreated, "application/json", resp)
	}
}
func handelePING(c *gin.Context) {
	if rnt.db != nil {
		err := rnt.db.Ping()
		if err != nil {
			c.Status(http.StatusInternalServerError)
		} else {
			c.Status(http.StatusOK)
		}
	} else {
		c.Status(http.StatusInternalServerError)
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
	r.GET("/ping", handelePING)
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
	rnt.keytoURLMap = make(map[string]string)
	ServerInit()
	defer rnt.db.Close()
	r := setupRouter()
	r.Run(cfg.ServerAddress)
}
