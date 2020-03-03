package redis

import (
	"encoding/json"
	"log"
	"os"

	"github.com/garyburd/redigo/redis"
	"github.com/jonas19/minesweeper/minesweeper/consts"
	"github.com/jonas19/minesweeper/minesweeper/models"
)

var client redis.Conn
var err error

type Services struct {
	GameService models.Game
}

func Start() {
	var redisURL string
	if os.Getenv("REDIS_URL_STUNNEL") != "" {
		redisURL = os.Getenv("REDIS_URL_STUNNEL")[5:]
	} else {
		redisURL = consts.RedisAddr
	}

	log.Println("Trying to connecto to redis on " + redisURL)
	client, err = redis.Dial("tcp", redisURL)

	if err != nil {
		panic("Unable to connect to Redis server")
	}
}

//save a game
func Persist(game models.Game) (status bool) {
	json, err := json.Marshal(game)
	if err != nil {
		return false
	}

	client.Do("SET", game.GameID, []byte(json))

	return true
}

//get game data
func LoadGame(gameID string) (status bool, data string) {
	data, err := redis.String(client.Do("GET", gameID))
	if err != nil {
		return false, ""
	}

	return true, data
}
