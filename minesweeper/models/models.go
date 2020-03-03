package models

type Cell struct {
	CellID      int
	IsAMine     bool
	IsClicked   bool
	IsFlagged   bool
	FlaggedWith string
}

type CellGrid []Cell

type Game struct {
	GameID        string
	Rows          int
	Cols          int
	CantCells     int
	RevealedCells int
	Mines         int
	Clicks        int
	Status        string
	Board         CellGrid
}
