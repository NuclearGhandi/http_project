package main

import (
	"flag"
	"math/rand"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(Logger())
	r.Use(gin.Recovery())
	r.Use(Gzip(DefaultCompression))

	r.GET("/", serverErr)
	r.GET("/ping", HandelePING)
	r.POST("/:key", serverErr)
	r.POST("/api/shorten", HandleAPIPOST)
	r.GET("/api/:key", serverErr)
	r.GET("/:key", HandleGET)
	r.POST("/", HandlePOST)
	r.POST("/api/shorten/batch", HandleBunch)
	return r
}

func serverErr(c *gin.Context) {
	c.AbortWithStatus(http.StatusBadRequest)
}
