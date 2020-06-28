package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// TODO: consider specifying buffer sizes
var upgrader = websocket.Upgrader{}

const projectID = "fir-test-9a9f3"

func main() {
	// Configure logging
	loadLogging()

	// Connect to db
	mongoConnectString := os.Getenv("MONGO_CONNECTION_URL")
	database = NewDB(mongoConnectString)

	// Connect to firebase
	fb = NewFirebase()

	// Create router
	r := gin.Default()

	// Establish profiling if desired
	if os.Getenv("PPROF") == "1" {
		pprof.Register(r)
	}

	// Configure templates and static files
	r.LoadHTMLFiles("client/public/index.html", "client/templates/test.html")
	r.Static("/js", "client/js")
	r.Static("/public", "client/public")
	r.Static("/static", "client/public/static")
	r.Static("/fonts", "client/public/fonts")
	r.StaticFile("/aubio.wasm", "client/public/aubio.wasm")
	r.StaticFile("/robots.txt", "client/robots.txt")
	r.StaticFile("/aubio.js", "client/public/aubio.js")
	r.StaticFile("/a.wasm", "client/public/a.wasm")
	r.StaticFile("/automation-icon-disabled.png", "client/public/automation-icon-disabled.png")
	r.StaticFile("/automation-icon-enabled.png", "client/public/automation-icon-enabled.png")
	r.StaticFile("/arrow-loop.png", "client/public/arrow-loop.png")
	r.StaticFile("/arrow-loop-looping.png", "client/public/arrow-loop-looping.png")
	r.StaticFile("/piano-icon.png", "client/public/piano-icon.png")
	r.StaticFile("/service-worker.js", "client/public/service-worker.js")
	r.StaticFile("/a.js", "client/public/a.js")
	r.StaticFile("/a.out.js", "client/public/a.out.js")
	r.StaticFile("/a.out.wasm", "client/public/a.out.wasm")
	r.StaticFile("/convolver/PrimeXtraLong.wav", "client/public/convolver/PrimeXtraLong.wav")
	r.StaticFile("/manifest.json", "client/public/manifest.json")
	r.StaticFile("/precache-manifest.50377630ce85b5db883d49016b4c961d.js", "client/public/precache-manifest.50377630ce85b5db883d49016b4c961d.js")

	// For pre-cache request (202 never 304)
	r.GET("/index.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Serve Chronophonics
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(func(c *gin.Context) {
		adminKey := c.Request.Header.Get("X-Admin-Key")
		if adminKey == "" {
			c.String(http.StatusUnauthorized, "authorization header not present")
			c.Abort()
			return
		}
		if adminKey != os.Getenv("ADMIN_KEY") {
			c.String(http.StatusUnauthorized, "authorization header invalid")
			c.Abort()
			return
		}
	})

	// Delete room operations
	admin.DELETE("rooms/:roomName/operations", func(c *gin.Context) {
		roomName := c.Param("roomName")
		err := database.DeleteAllOperations(roomName)
		if err != nil {
			c.String(http.StatusInternalServerError, "unable to delete all operations: %s", err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Delete all firebase users
	admin.DELETE("firebase/users", func(c *gin.Context) {
		err := fb.DeleteAllUsers()
		if err != nil {
			c.String(http.StatusInternalServerError, "unable to delete all users: %s", err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Turn on/off websocket CORS
	admin.POST("websocket/cors", func(c *gin.Context) {
		// Look for "enforce" query param
		enforceParam, ok := c.Request.URL.Query()["enforce"]
		if !ok {
			c.String(http.StatusBadRequest, "missing query param \"enforce\"")
			return
		}
		enforce := enforceParam[0]
		if strings.ToLower(enforce) == "false" {
			log.Infof("disabling CORS for websocket connections")
			upgrader.CheckOrigin = func(r *http.Request) bool {
				return true
			}
		} else {
			log.Infof("enabling CORS for websocket connections")
			upgrader.CheckOrigin = nil
		}
		c.Status(http.StatusAccepted)
	})

	// Websocket handler
	if env := os.Getenv("ENV"); env == "local" {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	r.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})

	r.Run(":80")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Unable to upgrade ws request: %s", err)
		return
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
