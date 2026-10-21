package main

import (
	"log"

	"github.com/Dhwanilv26/gopher-social/internal/db"
	"github.com/Dhwanilv26/gopher-social/internal/env"
	"github.com/Dhwanilv26/gopher-social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDRESS", "postgres://admin:adminpassword@localhost:5433/socialnetwork?sslmode=disable")
	
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store)
}
