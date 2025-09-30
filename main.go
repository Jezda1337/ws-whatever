package main

import (
	"fmt"
	"html/template"
	"log"
	"ws-whatever/ws"

	gws "github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	// in case of .env
	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASSWORD")
	// dbname := os.Getenv("DB_NAME")

	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "HgYKJ72T"
	dbname := "messaging"
	sslmode := "disable"

	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", host, user, password, dbname, port, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	// logger := utils.NewLogger()

	m := ws.NewManager(db)

	e.GET("/", func(c echo.Context) error {
		tmpl.Execute(c.Response().Writer, nil)
		return nil
	})

	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if err != nil {
			log.Printf("Something goes wrong with upgrading protocol: %v", err)
			return err
		}

		client := ws.NewClient(conn, m)
		client.Manager.AddClient(client)

		go client.ReadMessages()
		go client.WriteMessages()

		return nil
	})

	// serving static files
	e.Static("/static", "web/static")

	log.Println("Server running on port :6969")
	e.Logger.Fatal(e.Start(":6969"))
}
