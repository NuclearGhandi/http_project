package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func HandleUserReturn(c *gin.Context) {

}
func HandleGET(c *gin.Context) {
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
func HandlePOST(c *gin.Context) {
	if c.Param("key") != "" {
		serverErr(c)
	} else {
		body, err := c.GetRawData()
		user := CookieDecoder(c)
		url, dupl := addURL(string(body), user)
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
func HandleAPIPOST(c *gin.Context) {
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
	user := CookieDecoder(c)
	outpt.URL, flag = addURL(inpt.URL, user)
	resp, err := json.Marshal(outpt)
	if err != nil {
		serverErr(c)
	} else if flag {
		c.Data(http.StatusConflict, "application/json", resp)
	} else {
		c.Data(http.StatusCreated, "application/json", resp)
	}
}
func HandelePING(c *gin.Context) {
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
func HandleBunch(c *gin.Context) {
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
		user := CookieDecoder(c)
		outpt.URL, _ = addURL(inpt.URL, user)
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
