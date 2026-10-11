package main

import (
	"log"

	"github.com/Dhwanilv26/gopher-social/internal/env"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	cfg := config{
		addr: env.GetString("ADDRESS", ":8080"),
	}
	app := &application{config: cfg}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
