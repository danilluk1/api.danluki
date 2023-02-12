package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/danilluk1/api.danluki/db/entities"
)

type counterBody struct {
	NumRange int    `json:"numRange"`
	Path     string `json:"path"`
}

type counterResponse struct {
	Views int64 `json:"views"`
}

func (app *app) counterHandler() func(w http.ResponseWriter, r *http.Request) {
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

func dataHandler(w http.ResponseWriter, r *http.Request) {

}
