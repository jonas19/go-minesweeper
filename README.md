# Golang Minesweeper coding challenge

A minesweeper done in Golang

## Build and Run with Docker

docker-compose up

## API endpoints

* **127.0.0.1/** - [Get] Gets API version

* **127.0.0.1/healthcheck** - [Get] Perform a basic app health check

* **127.0.0.1/game** - [Post] Start a new game with default parameters

* **127.0.0.1/game/{gameID}/board** - [Post] Loads and retrieves game info

* **127.0.0.1/game/{gameID}/flag/{cellID}/question** - [Post] Flags a cell with a question mark or unflags 
it

* **127.0.0.1/game/{gameID}/flag/{cellID}/flag** - [Post] Flags a cell with a red flag or unflags it

* **127.0.0.1/game/{gameID}/click/{cellID}** - [Post] Clicks on a cell

## Run tests within Docker

docker exec -it app go test $(go list ./... ) -v

## Some thoughts

Instead of creating a matix of [rows][cols], I decided to use an array of rows*cols lenght.
I found it funny to experiment with this way of doing the app.

One of the challenges was to use the go routines in order to speed up the cell revealing proceess, specially for very big board.

## Unit testing

Done just a few unit testing just to show how will I do it in Go lang.
If requested, could be done to grant a 80% coverage, which I think is an acceptable threshold.

## Author

* **Jonas Garbovetsky** - *Initial work* - [Jonas19](https://github.com/jonas19)
