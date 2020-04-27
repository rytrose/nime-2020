package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{}

const projectID = "fir-test-9a9f3"

func main() {
	loadLogging()
	r := gin.Default()

	// Configure templates and static files
	r.LoadHTMLGlob("client/templates/*")
	r.Static("/js", "client/js")

	// Root app
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Wow, title!",
		})
	})

	// Websocket handler
	r.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})

	// Connect to db
	mongoConnectString := os.Getenv("MONGO_CONNECTION_URL")
	database = NewDB(mongoConnectString)

	r.Run(":80")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Unable to upgrade ws request: %s", err)
	}
	NewClient(conn)
}

func loadLogging() {
	l, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Warnf("Unable to parse LOG_LEVEL env var, defaulting to INFO: %s", err)
		l = log.InfoLevel
	}
	log.SetLevel(l)
	log.SetFormatter(&log.JSONFormatter{})
}
