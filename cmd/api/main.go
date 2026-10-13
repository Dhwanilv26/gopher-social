package main

import (
	"log"

	"github.com/Dhwanilv26/gopher-social/internal/db"
	"github.com/Dhwanilv26/gopher-social/internal/env"
	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	cfg := config{
		addr: env.GetString("ADDRESS", ":8080"),
		db: dBConfig{
			addr:         env.GetString("DB_ADDRESS", "postgres://admin:adminpassword@localhost:5433/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPENS_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("database connection pool established")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
