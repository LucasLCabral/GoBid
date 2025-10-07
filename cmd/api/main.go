package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/LucasLCabral/go-bid/internal/api"
	"github.com/LucasLCabral/go-bid/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("GOBID_DB_USER"),
		os.Getenv("GOBID_DB_PASSWORD"),
		os.Getenv("GOBID_DB_HOST"),
		os.Getenv("GOBID_DB_PORT"),
		os.Getenv("GOBID_DB_NAME"),
	))

	if err != nil {
		panic(err)
	}

	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	api := api.API{
		Router:      chi.NewMux(),
		UserService: services.NewUserService(pool),
	}
	api.BindRoutes()
	fmt.Println("Starting server on :3080")
	if err := http.ListenAndServe("localhost:3080", api.Router); err != nil {
		panic(err)
	}
}
