package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/danilluk1/api.danluki/db/entities"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	dbConn := os.Getenv("DB_CONN")
	if dbConn == "" {
		panic(errors.New("can't read variable DB_CONN"))
	}
	port := os.Getenv("PORT")
	if dbConn == "" {
		panic(errors.New("can't read variable PORT"))
	}

	db, err := gorm.Open(postgres.Open(dbConn))
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&entities.Statistics{})
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()
	router.Get("/data", dataHandler)
	router.Post("/counter", counterHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down...")
	os.Exit(0)
}
