package engine

import (
	"encoding/json"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

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

func (game *Services) GetJSONGameByID(gameID string) (status string, message string) {
	ok, message := game.obtainGameInfo(gameID)
	if !ok {
		return "error", message
	}

	s, _ := json.Marshal(game)
	
	return "ok", string(s)
}

func (game *Services) GetGraphicallyAGameByID(gameID string, mode string) (status string, message string) {
	ok, message := game.obtainGameInfo(gameID)
	if !ok {
		return "error", message
	}

	board := ""
	cols := 0
	for cell := range game.GameService.Board {
		if mode == "bombs" {
			if game.GameService.Board[cell].IsAMine {
					board += " * "
			} else {
				board += " - "
			}
		} else if mode == "current" {
			if game.GameService.Board[cell].FlaggedWith == "question" {
				board += " ? "
			} else if game.GameService.Board[cell].FlaggedWith == "flag" {
				board += " ! "
			} else if game.GameService.Board[cell].MinesSurrounding > 0 {
				board += "  " + strconv.Itoa(game.GameService.Board[cell].MinesSurrounding) + " "
			} else if game.GameService.Board[cell].IsAMine && game.GameService.Status == "Lost" {
				board += " * "
			} else if game.GameService.Board[cell].IsRevealed {
				board += " _ "
			} else {
				board += " - "
			}
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

	var cellsToReveal string
	if game.GameService.Board[cellID].IsAMine {
		game.GameService.Board[cellID].IsRevealed = true
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
	ok, message = game.obtainGameInfo(gameID)
	if !ok {
		return false, message
	}

	if game.GameService.Status != "playing" {
		return false, "Selected game is over"
	}

	ok, message = game.isCellOnRange(cellID)
	if !ok {
		return false, message
	}

	if game.GameService.Board[cellID].IsRevealed {
		return false, "Cell already revealed"
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
		game.GameService.Board[cell].CellID           = cell
		game.GameService.Board[cell].IsRevealed       = false
		game.GameService.Board[cell].IsFlagged        = false
		game.GameService.Board[cell].IsAMine          = false
		game.GameService.Board[cell].MinesSurrounding = 0
		game.GameService.Board[cell].Processed		  = false
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

func (game *Services) obtainGameInfo(gameID string) (ok bool, message string) {
	ok, data := redis.LoadGame(gameID)
	if !ok {
		return false, "Unable to load game"
	}

	ok = game.parseVariables(data)
	if !ok {
		return false, "Seems data was not properly saved, unable to load game"
	}

	return true, ""
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

	//was processed in this or other time
	if game.GameService.Board[cellID].Processed  {
		return
	}

	game.GameService.Board[cellID].Processed = true
	game.GameService.Board[cellID].IsRevealed = true 
	game.GameService.RevealedCells++

	//bring all adjacent cells
	adjacents := game.getAdjacentCells(cellID)
	
	for cell := range adjacents {
		if game.GameService.Board[adjacents[cell]].IsAMine {
			game.GameService.Board[cellID].MinesSurrounding++
		}
	}

	if game.GameService.Board[cellID].MinesSurrounding == 0 {
		lock.Lock()
		if !revealedCells[cellID] {
			revealedCells[cellID] = true
		}
		lock.Unlock()
	
		for cell := range adjacents {
			if game.GameService.Board[adjacents[cell]].IsRevealed == false && game.GameService.Board[adjacents[cell]].IsFlagged == false {
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

func (game *Services) getAdjacentCells(cellID int) (adjacentCells []int) {
	cellRowNumber   := math.Round(float64(cellID / game.GameService.Rows))
	topRowNumber    := cellRowNumber - 1
	bottomRowNumber := cellRowNumber + 1
	
	var targetTopRow []int
	targetTopRow = append(targetTopRow, (cellID-game.GameService.Rows)-1)
	targetTopRow = append(targetTopRow, (cellID-game.GameService.Rows))
	targetTopRow = append(targetTopRow, (cellID-game.GameService.Rows)+1)	
	
	for _, cell := range targetTopRow {
		whichRow := math.Round(float64(cell / game.GameService.Rows))
		if cell >= 0 && 
		   whichRow == topRowNumber {
			adjacentCells = append(adjacentCells, cell)
		}
	}

	var targetSameRow []int
	targetSameRow = append(targetSameRow, cellID-1)
	targetSameRow = append(targetSameRow, cellID+1)
	
	for _, cell := range targetSameRow {
		whichRow := math.Round(float64(cell / game.GameService.Rows))
		if cell > 0 &&
		   cell <  game.GameService.CantCells &&
		   whichRow == cellRowNumber {
			adjacentCells = append(adjacentCells, cell)
		}
	}

	var targetBottomRow []int
	targetBottomRow = append(targetBottomRow, (cellID+game.GameService.Rows)-1)
	targetBottomRow = append(targetBottomRow, (cellID+game.GameService.Rows))
	targetBottomRow = append(targetBottomRow, (cellID+game.GameService.Rows)+1)
	for _, cell := range targetBottomRow {
		whichRow := math.Round(float64(cell / game.GameService.Rows))
		if cell <  game.GameService.CantCells &&
		   whichRow == bottomRowNumber {
			adjacentCells = append(adjacentCells, cell)
		}
	}

	return adjacentCells
}
