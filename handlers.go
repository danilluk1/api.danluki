package main

import "net/http"

type counterBody struct {
	NumRange int    `json:"numRange"`
	Path     string `json:"path"`
}

func counterHandler(w http.ResponseWriter, r *http.Request) {
	body := r.Body.Read()
}

func dataHandler(w http.ResponseWriter, r *http.Request) {

}
