package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"ws-whatever/ws"

	gws "github.com/gorilla/websocket"
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

	mux := http.NewServeMux()
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	// logger := utils.NewLogger()

	m := ws.NewManager(db)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Something goes wrong with upgrading protocol: %v", err)
			return
		}

		c := ws.NewClient(conn, m)
		c.Manager.AddClient(c)

		go c.ReadMessages()
		go c.WriteMessages()
	})

	// serving static files
	fs := http.FileServer(http.Dir("web/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	log.Println("Server running on port :6969")
	log.Fatal(http.ListenAndServe(":6969", mux))
}
