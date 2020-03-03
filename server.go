package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"github.com/minesweeper/routes"
	"github.com/minesweeper/consts"
)

func main() {
	r := mux.NewRouter()
	log.Println("Minesweeper running on port " + consts.Port)

	//service := routes.Services{}
	//, consts.Default_cols, consts.Default_mines, consts.Default_board
	
	r.HandleFunc("/game",                		        routes.StartNewGame).Methods(http.MethodPost)
	r.HandleFunc("/game/{gameID}/board", 		        routes.RetrieveGame).Methods(http.MethodGet)
	r.HandleFunc("/game/{gameID}/flag/{cellID}/{with}", routes.FlagCell).Methods(http.MethodPost)
	r.HandleFunc("/game/{gameID}/click/{cellID}",       routes.ClickCell).Methods(http.MethodPost)
	r.HandleFunc("/",                    		        routes.APIVersion).Methods(http.MethodGet)
	
	log.Fatalln(http.ListenAndServe(":" + consts.Port, r))
}