package redis

import (
	"encoding/json"
	"os"

	"github.com/garyburd/redigo/redis"
	"github.com/jonas19/minesweeper/minesweeper/consts"
	"github.com/jonas19/minesweeper/minesweeper/models"
	"github.com/sirupsen/logrus"
)

var client redis.Conn
var err error

type Services struct {
	GameService models.Game
}

func Start() {
	var redisURL string

	if os.Getenv("REDIS_URL") != "" {
		redisURL = os.Getenv("REDIS_URL")
	} else if os.Getenv("REDIS_URL_STUNNEL") != "" {
		redisURL = os.Getenv("REDIS_URL_STUNNEL")[5:]
	} else {
		redisURL = consts.RedisAddr
	}

	client, err = redis.Dial("tcp", redisURL)
	if err != nil {
		panic("Unable to connect to Redis server")
	}
}

//save a game
func Persist(game models.Game) (status bool) {
	Start()

	json, err := json.Marshal(game)
	if err != nil {
		return false
	}

	log := logrus.StandardLogger()
	log.Infoln("Persisting " + game.GameID)
	log.Infoln(string(json))
	client.Do("SET", game.GameID, string(json))

	return true
}

//get game data
func LoadGame(gameID string) (status bool, data string) {
	Start()

	data, err := redis.String(client.Do("GET", gameID))

	log := logrus.StandardLogger()
	log.Infoln("Loading " + gameID)

	if err != nil {
		log.Infoln("error!!")
		log.Infoln(err)
		return false, ""
	} else {
		log.Infoln(data)
	}

	return true, data
}
