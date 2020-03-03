package engine

import (
	"encoding/json"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/jonas19/minesweeper/minesweeper/consts"
	"github.com/jonas19/minesweeper/minesweeper/models"
	"github.com/jonas19/minesweeper/minesweeper/storage/redis"
)

var lock sync.RWMutex

type Services struct {
	GameService models.Game
}

func init() {
	rand.Seed(time.Now().Unix())
}

func (game *Services) StartANewGame(rows string, cols string, mines string) (status string, message string) {
	var defaultForMines int

	ok, err := game.setParameters("rows",
		&game.GameService.Rows,
		consts.Default_rows,
		rows,
		1,
		255)
	if !ok {
		return "error", err
	}

	ok, err = game.setParameters("cols",
		&game.GameService.Cols,
		consts.Default_cols,
		cols,
		1,
		255)
	if !ok {
		return "error", err
	}

	game.GameService.CantCells = game.GameService.Cols * game.GameService.Rows

	if consts.Default_mines < game.GameService.CantCells {
		defaultForMines = consts.Default_mines
	} else {
		defaultForMines = int(math.Round(float64(game.GameService.CantCells) * 0.2))
	}

	ok, err = game.setParameters("mines",
		&game.GameService.Mines,
		defaultForMines,
		mines,
		1,
		game.GameService.CantCells)
	if !ok {
		return "error", err
	}

	//create the board
	game.createBoard()

	//randomly distribute the mines
	game.randomlyDistributeMines()

	//set the game status playing
	game.GameService.Status = "playing"

	//no cell has been revealed yet
	game.GameService.RevealedCells = 0

	//generate a random game id to identify the game
	game.GameService.GameID = strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + strconv.Itoa(rand.Intn(10000))

	//save the game to the db
	ok = redis.Persist(game.GameService)
	if !ok {
		return "error", "Unable to save the game"
	}

	return "ok", game.GameService.GameID
}

func (game *Services) GetGraphicallyAGameByID(gameID string) (status string, message string) {
	ok := game.obtainGameInfo(gameID)
	if !ok {
		return "error", "Seems data was not properly saved, unable to load game"
	}

	board := ""
	cols := 0
	for cell := range game.GameService.Board {
		if game.GameService.Board[cell].FlaggedWith == "question" {
			board += " ? "
		} else if game.GameService.Board[cell].FlaggedWith == "flag" {
			board += " ! "
		} else if game.GameService.Board[cell].IsClicked {
			board += " " + strconv.Itoa(game.GameService.Board[cell].MinesSorrounding) + " "
		} else {
			board += " - "
		}

		cols++
		if cols == game.GameService.Cols {
			board += "\n"
			cols = 0
		}
	}

	return status, board
}

func (game *Services) FlagACell(gameID string, cellIDstr string, with string) (status string, message string) {
	cellID, err := strconv.Atoi(cellIDstr)
	if err != nil {
		return "error", "Unable to process cellID"
	}

	ok, message := game.basicChekUpBeforePlaying(gameID, cellID)
	if !ok {
		return "error", message
	}

	if game.GameService.Board[cellID].IsFlagged {
		message = "Cell unflagged"
		game.GameService.Board[cellID].FlaggedWith = ""
	} else {
		if with == "question" {
			game.GameService.Board[cellID].FlaggedWith = "question"
			message = "Cell flagged with a question mark"
		} else {
			game.GameService.Board[cellID].FlaggedWith = "flag"
			message = "Cell red flagged"
		}
	}

	game.GameService.Board[cellID].IsFlagged = !game.GameService.Board[cellID].IsFlagged

	ok = redis.Persist(game.GameService)
	if !ok {
		return "error", "Unable to save the game"
	}

	return "ok", message
}

func (game *Services) ClickACell(gameID string, cellIDstr string) (status string, message string) {
	cellID, err := strconv.Atoi(cellIDstr)
	if err != nil {
		return "error", "Unable to process cellID"
	}

	ok, message := game.basicChekUpBeforePlaying(gameID, cellID)
	if !ok {
		return "error", message
	}

	//game.GameService.Board[cellID].IsClicked = true
	//game.GameService.Clicks++

	var cellsToReveal string
	if game.GameService.Board[cellID].IsAMine {
		game.GameService.Board[cellID].IsClicked = true
		game.GameService.Clicks++
		game.GameService.Status = "Lost"
	} else {
		var wg sync.WaitGroup
		cellsToRevealArr := make(map[int]bool)

		cellsToRevealArr[cellID] = true

		wg.Add(1)
		game.revealAdjacent(cellID, cellsToRevealArr, &wg)
		wg.Wait()
		s, _ := json.Marshal(cellsToRevealArr)
		cellsToReveal = strings.Trim(string(s), "[]")

		game.GameService.Board[cellID].IsClicked = true
		game.GameService.Clicks++
	}

	game.checkIfWon()

	ok = redis.Persist(game.GameService)
	if !ok {
		return "error", "Unable to save game"
	}

	if game.GameService.Status == "Lost" {
		return "ok", "Sorry, you lost!"
	} else if game.GameService.Status == "Won" {
		return "ok", "Win! Win! Win!"
	} else {
		return "ok", "cellsToReveal: [" + cellsToReveal + "]"
	}

}

func (game *Services) basicChekUpBeforePlaying(gameID string, cellID int) (ok bool, message string) {
	ok = game.obtainGameInfo(gameID)
	if !ok {
		return false, "Seems data was not properly saved, unable to load game"
	}

	if game.GameService.Status != "playing" {
		return false, "Selected game is over"
	}

	ok, message = game.isCellOnRange(cellID)
	if !ok {
		return false, message
	}

	if game.GameService.Board[cellID].IsClicked {
		return false, "Cell already clicked"
	}

	return true, ""
}

func (game *Services) setParameters(name string,
	paramToSet *int,
	defaultValue int,
	value string,
	min int,
	max int) (ok bool, response string) {

	if value == "" {
		*paramToSet = defaultValue
	} else {
		if valInt, err := strconv.Atoi(value); err == nil {
			if valInt >= min && valInt <= max {
				*paramToSet = int(valInt)
			} else {
				return false, "Invalid " + name + " value, expected " + strconv.Itoa(min) + "-" + strconv.Itoa(max)
			}
		} else {
			return false, "Invalid " + name + " value, expected integer between " + strconv.Itoa(min) + "-" + strconv.Itoa(max)
		}
	}

	return true, ""
}

func (game *Services) createBoard() {
	game.GameService.Board = make(models.CellGrid, game.GameService.CantCells)

	for cell := range game.GameService.Board {
		game.GameService.Board[cell].CellID = cell
		game.GameService.Board[cell].IsClicked = false
		game.GameService.Board[cell].IsFlagged = false
		game.GameService.Board[cell].IsAMine = false
		game.GameService.Board[cell].MinesSorrounding = 0
	}
}

func (game *Services) randomlyDistributeMines() {
	bombsLeftToSet := game.GameService.Mines
	for bombsLeftToSet > 0 {
		idx := rand.Intn(game.GameService.CantCells)
		if !game.GameService.Board[idx].IsAMine {
			game.GameService.Board[idx].IsAMine = true
			bombsLeftToSet--
		}
	}
}

func (game *Services) parseVariables(dataToParse string) (ok bool) {
	if err := json.Unmarshal([]byte(dataToParse), &game.GameService); err != nil {
		return false
	}

	return true
}

func (game *Services) obtainGameInfo(gameID string) (ok bool) {
	ok, data := redis.LoadGame(gameID)
	if !ok {
		return false
	}

	ok = game.parseVariables(data)
	if !ok {
		return false
	}

	return true
}

func (game *Services) isCellOnRange(cellID int) (ok bool, response string) {
	if cellID >= game.GameService.CantCells {
		return false, "CellID should be less than " + strconv.Itoa(game.GameService.CantCells-1)
	}

	if cellID < 0 {
		return false, "CellID should be greater than 0"
	}

	return true, ""
}

func (game *Services) revealAdjacent(cellID int, revealedCells map[int]bool, wg *sync.WaitGroup) {
	defer wg.Done()
	log := logrus.StandardLogger()
	log.Infoln("Check A " + strconv.Itoa(cellID))

	if game.GameService.Board[cellID].IsClicked || game.GameService.Board[cellID].IsFlagged {
		return
	}

	//bring all adjacent cells
	adjacents := game.getAdjacentCells(cellID)

	for cell := range adjacents {
		if game.GameService.Board[adjacents[cell]].IsAMine {
			game.GameService.Board[cellID].MinesSorrounding++
			log.Infoln("mina alrededor")
		}
	}

	log.Infoln("MinesSorrounding " + strconv.Itoa(game.GameService.Board[cellID].MinesSorrounding))
	if game.GameService.Board[cellID].MinesSorrounding == 0 {
		lock.Lock()
		if !revealedCells[cellID] {
			revealedCells[cellID] = true
		}
		lock.Unlock()
		for cell := range adjacents {
			log.Infoln("Check B" + strconv.Itoa(adjacents[cell]))
			if game.GameService.Board[adjacents[cell]].IsClicked == false || game.GameService.Board[adjacents[cell]].IsFlagged == false {
				game.GameService.Board[adjacents[cell]].IsClicked = true
				game.GameService.RevealedCells++
				wg.Add(1)
				go game.revealAdjacent(adjacents[cell], revealedCells, wg)
			}
		}
	}
}

func (game *Services) checkIfWon() {
	if game.GameService.Clicks+game.GameService.RevealedCells+game.GameService.Mines == game.GameService.CantCells {
		game.GameService.Status = "Won"
	}
}

func (game *Services) getAdjacentCells(cellID int) (adjacents []int) {
	OriginalRowNumber := math.Round(float64(cellID / game.GameService.Rows))

	var cellToCheckIfIsAdjacent []int

	//NW
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID-game.GameService.Rows)-1)
	//N
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID - game.GameService.Rows))
	//NE
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID-game.GameService.Rows)+1)
	//W
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID - 1))
	//E
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID + 1))
	//SW
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID+game.GameService.Rows)-1)
	//S
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID + game.GameService.Rows))
	//SE
	cellToCheckIfIsAdjacent = append(cellToCheckIfIsAdjacent, (cellID+game.GameService.Rows)+1)

	for _, value := range cellToCheckIfIsAdjacent {
		if value >= 0 &&
			value < game.GameService.CantCells &&
			(OriginalRowNumber-1) == math.Round(float64(value/game.GameService.Rows)) {
			adjacents = append(adjacents, value)
		}
	}

	return adjacents
}
