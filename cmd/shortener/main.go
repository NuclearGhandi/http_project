package main

import (
	"database/sql"

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

var cfg Config
var rnt Runtime

func main() {
	rnt.keytoURLMap = make(map[string]string)
	ServerInit()
	defer rnt.db.Close()
	r := SetupRouter()
	r.Run(cfg.ServerAddress)
}
