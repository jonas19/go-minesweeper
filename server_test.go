package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
 	"github.com/gorilla/mux"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/minesweeper/types"
	"github.com/minesweeper/routes"	
)

type Services struct {
	GameService types.Game
}

func Router() *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/game",                		        routes.StartNewGame).Methods(http.MethodPost)
	r.HandleFunc("/game/{gameID}/board", 		        routes.RetrieveGame).Methods(http.MethodGet)
	r.HandleFunc("/game/{gameID}/flag/{cellID}/{with}", routes.FlagCell).Methods(http.MethodPost)
	r.HandleFunc("/game/{gameID}/click/{cellID}",       routes.ClickCell).Methods(http.MethodPost)
	r.HandleFunc("/",                    		        routes.APIVersion).Methods(http.MethodGet)
	
    return r
}

func TestServerIsRunning(t *testing.T) {
 	request, _ := http.NewRequest("GET", "/", nil)
    response := httptest.NewRecorder()
    Router().ServeHTTP(response, request)
    assert.Equal(t, 200, response.Code, "OK response is expected")
}

func TestDefaultGameConfiguration(t *testing.T) {
 	request, _ := http.NewRequest("POST", "/game", nil)
    response := httptest.NewRecorder()   	
    Router().ServeHTTP(response, request)
    assert.Equal(t, 
    			 200, 
    			 response.Code,
    			 "OK response is expected")
    
    body, _ := ioutil.ReadAll(response.Body)
    assert.Contains(t, 
    			    string(body),
    			    "\"rows\": \"32\",\"cols\": \"32\",\"mines\": \"14\",}",    			    
    			    "Wrong default response from server")
}