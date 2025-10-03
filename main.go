package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"ws-whatever/internal"
	"ws-whatever/internal/db"
	"ws-whatever/utils"
	"ws-whatever/ws"

	gws "github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func testAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIDStr := c.QueryParam("user_id")
		if userIDStr == "" {
			userIDStr = string(time.Now().Second())
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			return echo.NewHTTPError(400, "invalid user_id")
		}

		c.Set("user_id", userID)
		return next(c)
	}
}

func main() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "HgYKJ72T")
	dbname := getEnv("DB_NAME", "messaging")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", host, user, password, dbname, port, sslmode)
	dbClient, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.RunMigration(dbClient); err != nil {
		log.Printf("Run migrations failed: %v", err)
	}

	e := echo.New()
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	logger := utils.NewLogger()

	m := ws.NewManager(dbClient, logger)

	e.GET("/", func(c echo.Context) error {
		if err := tmpl.Execute(c.Response(), nil); err != nil {
			log.Printf("Template execution error: %v", err)
			return err
		}
		return nil
	})

	e.GET("/ws", func(c echo.Context) error {
		userID := c.Get("user_id")
		if userID == nil {
			return echo.NewHTTPError(401, "unauthorized")
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return err
		}

		client := ws.NewClient(conn, m, userID.(int))
		client.Manager.AddClient(client)

		go client.ReadMessages()
		go client.WriteMessages()

		return nil
	}, testAuthMiddleware)

	// HTTP REST endpoints
	e.POST("/rooms", internal.CreateRoom(dbClient))
	e.GET("/rooms", internal.ListRooms(dbClient))
	e.GET("/rooms/:id/messages", internal.GetRoomMessages(dbClient))
	e.POST("/rooms/:id/participants", internal.AddRoomParticipant(dbClient))
	e.GET("/users/rooms", internal.GetUserRooms(dbClient), testAuthMiddleware)
	e.POST("/direct-messages", internal.CreateOrGetDirectMessage(dbClient), testAuthMiddleware)
	e.DELETE("/messages/:id", internal.DeleteMessage(dbClient), testAuthMiddleware)
	e.GET("/search/messages", internal.SearchMessages(dbClient))

	// serving static files
	e.Static("/static", "web/static")

	log.Println("Server running on port :6969")
	e.Logger.Fatal(e.Start(":6969"))
}
