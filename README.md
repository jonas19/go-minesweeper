# Goland Minesweeper coding challenge

A super simple minesweeper done in Golang

## API endpoints

* **/** - Gets API version

* **/game** - Start a new game with default parameters

* **/game/{gameID}/board** - Loads and retrieves game info

* **/game/{gameID}/flag/{cellID}/question** - Flags a cell with a question mark or unflags 
it

* **/game/{gameID}/flag/{cellID}/flag** - Flags a cell with a red flag or unflags it

* **/game/{gameID}/click/{cellID}** - Clicks on a cell

## Documentation

The code itself is quite self explanatory; clear and mnemotechnic methods and variables names were use.

## Persistency

In a true world app, Redis DB should be use in order to achive speed.
For the sake of simplicity for this project, game data is written to a file.

## GO routines

When starting a new game and definig the size of the board, go routines where used in order to show how they can be implemented in order to make concurrent proccesing with Go.    

## Testing

Some simple unit test were implemented in order to show how it should be done in Go

## Author

* **Jonas Garbovetsky** - *Initial work* - [PurpleBooth](https://github.com/jonas19)
