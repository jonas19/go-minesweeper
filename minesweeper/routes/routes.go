package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jonas19/minesweeper/minesweeper/consts"
	"github.com/jonas19/minesweeper/minesweeper/engine"
	"github.com/jonas19/minesweeper/minesweeper/httpresponses"
)

var game = engine.Services{}

func ShowVersion(w http.ResponseWriter, r *http.Request) {
	httpresponses.SendResponse(w, "ok", "App version: "+consts.API_Version)
}

func StartNewGame(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	status, message := game.StartANewGame(
		r.FormValue("rows"),
		r.FormValue("cols"),
		r.FormValue("mines"),
	)

	httpresponses.SendResponse(w, status, message)
}

func RetrieveGame(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)

	status, message := game.GetAGameByID(query["gameID"])

	httpresponses.SendResponse(w, status, message)
}

func FlagCell(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)

	status, message := game.FlagACell(query["gameID"], query["cellID"], query["with"])

	httpresponses.SendResponse(w, status, message)
}

func ClickCell(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)

	status, message := game.ClickACell(query["gameID"], query["cellID"])

	httpresponses.SendResponse(w, status, message)
}
