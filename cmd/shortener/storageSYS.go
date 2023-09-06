package main

import (
	"bufio"
	"encoding/json"
	"fmt"

	"database/sql"
	"math/rand"
	"os"

	_ "github.com/lib/pq"
)

type fileJSON struct {
	UUID        int    `json:"uuid,string"`
	ShortURL    string `json:"short_url,string"`
	OriginalURL string `json:"original_url,string"`
}

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
	var id int
	row := rnt.db.QueryRow("INSERT INTO shorted (id, seq, url) VALUES ($1, $2, $3) ON CONFLICT (url) DO UPDATE SET url = $3 RETURNING id", rnt.dbID, key, url)
	err := row.Scan(&id)
	if err != nil {
		rnt.sugar.Errorw(err.Error(), "event", "dbRead")
	}
	fmt.Println(key, url)
	fmt.Println(dbReadURL(key))
	if int(id) == rnt.dbID {
		rnt.dbID = rnt.dbID + 1
		return key, false
	} else {
		row := rnt.db.QueryRow(
			"SELECT seq FROM shorted WHERE id = $1", id)
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
	var id int
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
		sqlStatment := `INSERT INTO shorted (id, seq, url) VALUES ($1, $2, $3) ON CONFLICT (url) DO UPDATE SET seq = EXCLUDED.seq RETURNING id`
		row := rnt.db.QueryRow(sqlStatment, rnt.dbID, buf.ShortURL, buf.OriginalURL)
		err := row.Scan(&id)
		if err != nil {
			rnt.sugar.Errorw(err.Error(), "event", "dbRead")
		}
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
