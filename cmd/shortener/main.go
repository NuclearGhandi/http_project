package main

import (
	"bufio"
	"bytes"
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
	dbID        int
}

type fileJSON struct {
	UUID        int    `json:"uuid,string"`
	ShortURL    string `json:"short_url,string"`
	OriginalURL string `json:"original_url,string"`
}

type inputJSON struct {
	URL string `json:"url"`
}
type inputBunchJSON struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}
type outputBunchJSON struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}
type outputJSON struct {
	URL string `json:"result"`
}

var cfg Config
var rnt Runtime

func DatabaseInit() {
	var buff int
	buf, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "databaseInit")
	}
	rnt.db = buf
	fmt.Println("DB Init")
	if err != nil {
		rnt.sugar.Fatalw(err.Error(), "event", "databaseInit")
	}
	_, errr := rnt.db.Exec("CREATE TABLE IF NOT EXISTS shorted (id INTEGER PRIMARY KEY, seq VARCHAR(10), url VARCHAR(2084) UNIQUE)")
	if errr != nil {
		rnt.sugar.Errorw(errr.Error(), "event", "dbInit")
	}
	row := rnt.db.QueryRow("SELECT MAX(id) FROM shorted")
	err = row.Scan(&buff)
	rnt.dbID = buff + 1
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "databaseInit")
	}
}

func dbWriteURL(key string, url string) (string, bool) {
	rtrn, err := rnt.db.Exec("INSERT INTO shorted (id, seq, url) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id", rnt.dbID, key, url)
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "dbWrite")
	}
	id, err := rtrn.LastInsertId()
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "dbWrite")
	}
	fmt.Println(key, url)
	fmt.Println(dbReadURL(key))
	if int(id) == rnt.dbID {
		rnt.dbID = rnt.dbID + 1
		return key, false
	} else {
		row := rnt.db.QueryRow(
			"SELECT URL FROM shorted WHERE id = $1", id)
		err := row.Scan(&key)
		if err != nil {
			rnt.sugar.Errorw(err.Error(), "event", "dbWrite")
		}
		return key, true
	}

}

func dbFmt() {
	var buf int
	var URL string
	var longURL string
	var maxID int
	var minID int
	row := rnt.db.QueryRow("SELECT MAX(id) FROM shorted")
	err := row.Scan(&buf)
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "dbRead")
	}
	maxID = buf
	row = rnt.db.QueryRow("SELECT MIN(id) FROM shorted")
	err = row.Scan(&buf)
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "dbRead")
	}
	minID = buf
	for i := minID; i <= maxID; i = i + 1 {
		row = rnt.db.QueryRow("SELECT seq FROM shorted WHERE id = $1", i)
		err := row.Scan(&URL)
		if err != nil {
			rnt.sugar.Errorw(err.Error(), "event", "dbRead")
		}
		row = rnt.db.QueryRow("SELECT url FROM shorted WHERE id = $1", i)
		err = row.Scan(&longURL)
		if err != nil {
			rnt.sugar.Errorw(err.Error(), "event", "dbRead")
		}
		fmt.Println(i, URL, longURL)
	}
}
func dbReadURL(key string) string {
	var url string
	row := rnt.db.QueryRow(
		"SELECT URL FROM shorted WHERE seq = $1", key)
	err := row.Scan(&url)
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "dbRead")
	}
	return url
}

func FileDBTransfer() {
	var file *os.File
	var scanner *bufio.Scanner
	var err error
	var buf fileJSON
	rnt.fileLen = 0
	file, err = os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "FileReadOpen")
	}
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()
		if err = json.Unmarshal(data, &buf); err != nil {
			rnt.sugar.Fatalw(err.Error(), "event", "FileReadMarshalErr")
		}
		rnt.fileLen = rnt.fileLen + 1
		fmt.Println(rnt.dbID, buf.ShortURL, buf.OriginalURL)
		sqlStatment := `INSERT INTO shorted (id, seq, url) VALUES ($1, $2, $3) ON CONFLICT (url) DO UPDATE SET seq = $2 RETURNING id`
		rsp, err := rnt.db.Exec(sqlStatment, rnt.dbID, buf.ShortURL, buf.OriginalURL)

		if err != nil {
			rnt.sugar.Errorw(err.Error(), "event", "dbWrite")
		}
		id, err := rsp.LastInsertId()
		if int(id) == rnt.dbID {
			rnt.dbID = rnt.dbID + 1
		}
		if err != nil {
			rnt.sugar.Errorw(err.Error(), "event", "dbWrite")
		}
	}
	file.Close()
	dbFmt()
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
	//fmt.Println(cfg.DatabaseDSN)
	//fmt.Println(cfg.typeOfStorage)
	//fmt.Println(cfg.FileStoragePath)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func addURL(url string) (string, bool) {
	flag := false
	key := randSeq(8)

	if cfg.typeOfStorage == "file" || cfg.typeOfStorage == "map" {
		rnt.keytoURLMap[key] = url
	}
	if cfg.typeOfStorage == "file" {
		FileWrite(key, url)
	}
	if cfg.typeOfStorage == "db" {
		//fmt.Println("db write")
		key, flag = dbWriteURL(key, url)

	}
	outURL := cfg.BaseURL + "/" + key
	return outURL, flag
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
		url, dupl := addURL(string(body))

		if err != nil {
			serverErr(c)
		}
		if dupl {
			c.String(http.StatusConflict, url)
		} else {
			c.String(http.StatusCreated, url)
		}
	}
}
func handleAPIPOST(c *gin.Context) {
	var inpt inputJSON
	var outpt outputJSON
	var flag bool
	body, err := c.GetRawData()
	if err != nil {
		serverErr(c)
	}
	if err = json.Unmarshal(body, &inpt); err != nil {
		serverErr(c)
	}
	outpt.URL, flag = addURL(inpt.URL)
	resp, err := json.Marshal(outpt)
	if err != nil {
		serverErr(c)
	} else if flag {
		c.Data(http.StatusConflict, "application/json", resp)
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
func handleBunch(c *gin.Context) {
	var inpt inputBunchJSON
	var outpt outputBunchJSON
	var buf []byte
	var resp []byte
	var err error

	resp = append(resp, byte('['))
	body, err := c.GetRawData()
	if err != nil {
		serverErr(c)
	}
	found := true
	fmt.Println(string(body))
	body = body[1 : len(body)-2]
	//fmt.Println(string(body))
	for found {
		buf, body, found = bytes.Cut(body, []byte("}"))
		if bytes.IndexAny(buf, ",") == 0 {
			buf = buf[1:]
		}
		fmt.Println(string(append(buf, byte('}'))))
		fmt.Println("_________\n_________")
		if err = json.Unmarshal(append(buf, byte('}')), &inpt); err != nil {
			rnt.sugar.Fatalw(err.Error(), "event", "FileReadMarshalErr")
		}
		outpt.ID = inpt.ID
		outpt.URL, _ = addURL(inpt.URL)
		buff, err := json.Marshal(outpt)
		resp = append(resp, buff...)
		resp = append(resp, byte(','), byte('\n'))
		if err != nil {
			serverErr(c)
		}
	}
	resp = append(resp[:len(resp)-2], byte(']'))
	c.Data(http.StatusCreated, "application/json", resp)
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

	r.GET("/", serverErr)
	r.GET("/ping", handelePING)
	r.POST("/:key", serverErr)
	r.POST("/api/shorten", handleAPIPOST)
	r.GET("/api/:key", serverErr)
	r.GET("/:key", handleGET)
	r.POST("/", handlePOST)
	r.POST("/api/shorten/batch", handleBunch)
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
