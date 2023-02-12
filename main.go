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

	"encoding/json"
	"io"

	"github.com/danilluk1/api.danluki/db/entities"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	DB *gorm.DB
}

type counterBody struct {
	NumRange int    `json:"numRange"`
	Path     string `json:"path"`
}

type counterResponse struct {
	Views int64 `json:"views"`
}

func (app *App) counterHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can't read body", http.StatusBadRequest)
			return
		}

		var dto counterBody
		if err := json.Unmarshal(body, &dto); err != nil {
			log.Println(err)
			http.Error(w, "Error in body format", http.StatusBadRequest)
			return
		}

		if dto.NumRange < 1 || dto.NumRange == 0 || dto.Path == "" {
			log.Println("Incorrect data.")
			http.Error(w, "Incorrect data.", http.StatusBadRequest)
			return
		}

		if dto.NumRange != 1 {
			var findReq = entities.Statistics{
				Path:     dto.Path,
				NumRange: dto.NumRange,
			}
			app.DB.FirstOrCreate(&findReq)

			currentTimeMs := time.Now().UnixMilli()
			timeFromDbMs := findReq.RequestedAt.UnixMilli()

			if timeFromDbMs < currentTimeMs && findReq.Path == dto.Path {
				app.DB.Create(&entities.Statistics{
					Path:     dto.Path,
					NumRange: dto.NumRange,
				})
			}
		} else {
			app.DB.Create(&entities.Statistics{
				Path:     dto.Path,
				NumRange: dto.NumRange,
			})
		}

		var views int64
		app.DB.Model(&entities.Statistics{}).Where("path = ?", dto.Path).Count(&views)

		data, err := json.Marshal(counterResponse{
			Views: views,
		})
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error.", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

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

	app := App{
		DB: db,
	}

	router := chi.NewRouter()
	router.Post("/counter", app.counterHandler())

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Println(err)
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
