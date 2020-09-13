package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jonas19/minesweeper/minesweeper/consts"
	"github.com/jonas19/minesweeper/minesweeper/routes"
	"github.com/jonas19/minesweeper/minesweeper/storage/redis"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func main() {
	log = logrus.StandardLogger()

	redis.Start()

	var usePort string
	flag.StringVar(&usePort, "port", "", "port in which the app runs")
	flag.Parse()

	//if no port is defined via command line, take the default port
	if usePort == "" {
		if os.Getenv("PORT") != "" {
			log.Infoln("On Heroku, setting port to " + os.Getenv("PORT"))
			usePort = os.Getenv("PORT")
		} else {
			log.Infoln("Setting default port to " + consts.Port)
			usePort = consts.Port
		}
	} else {
		log.Infoln("Setting port to " + usePort)
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	//new game endpoint
	r.HandleFunc("/game", routes.StartNewGame).Methods(http.MethodPost)

	//retrieve current status of a game
	r.HandleFunc("/game/{gameID}/board/status", routes.RetrieveGameCurrentStatus).Methods(http.MethodGet)

	//retrieve bomb location of a game
	r.HandleFunc("/game/{gameID}/board/bombs", routes.RetrieveGameBombs).Methods(http.MethodGet)

	//retrieve save a game in JSON
	r.HandleFunc("/game/{gameID}/board/json", routes.RetrieveGameJSON).Methods(http.MethodGet)

	//flag/unflag a cell on saved gameID game
	r.HandleFunc("/game/{gameID}/flag/{cellID}/{with}", routes.FlagCell).Methods(http.MethodPost)

	//click a cell on saved gameID game
	r.HandleFunc("/game/{gameID}/click/{cellID}", routes.ClickCell).Methods(http.MethodPost)

	//get current app version
	r.HandleFunc("/", routes.ShowVersion).Methods(http.MethodGet)

	log.Infoln("Starting API...")
	log.Fatalln(http.ListenAndServe(":"+usePort, r))
}

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infoln(r.RemoteAddr, r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}
