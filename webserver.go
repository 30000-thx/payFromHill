package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func webserver() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8844"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index_new.tmpl.html", nil)
	})
	router.POST("/start", func(c *gin.Context) {
		acc := c.PostForm("accountNumber")
		if acc == "" {
			c.JSON(500, gin.H{
				"message": "unknown accountNumber",
			})
			return
		}
		total, err := strconv.ParseFloat(c.PostForm("total"), 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "unknown total",
				"err":     err,
			})
			return
		}

		t := time.Now()
		pa := PayAction{
			StartDate: t,
			ID:        t.Format("20060102150405"),
			// IsDebug:       true,
			AccountNumber: acc,
			//dollars to cents
			Total: int(total * 100),
		}

		err = setPayAction(&pa)

		if err != nil {
			c.JSON(500, gin.H{
				"result":  false,
				"message": err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"result":  true,
				"message": "ok",
			})
		}
	})

	router.GET("/currentPa", func(c *gin.Context) {
		c.JSON(200, currentPa)
	})

	router.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})

	router.GET("/getPayActions", func(c *gin.Context) {
		actions, err := GetPayActions()
		if err != nil {
			c.JSON(500, gin.H{
				"message": "error",
				"err":     err,
			})
			return
		}
		c.JSON(200, actions)
	})

	router.GET("/getStatus", func(c *gin.Context) {
		c.JSON(200, status)
	})

	router.GET("/startPayment", func(c *gin.Context) {
		go startPay()
		c.JSON(200, gin.H{
			"result":  true,
			"message": "ok",
		})
	})

	go router.Run(":" + port)

	openbrowser("http://127.0.0.1:" + port)
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var conn *websocket.Conn

func wshandler(w http.ResponseWriter, r *http.Request) {
	c, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}

	conn = c

	sendMsgToWs(pa)
}

func sendMsgToWs(msg interface{}) {
	if conn == nil {
		return
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	fmt.Println("sending msg", string(b))
	conn.WriteMessage(websocket.TextMessage, b)
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
