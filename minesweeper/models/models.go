package models

type Cell struct {
	CellID           int    `json:"CellID"`
	IsAMine          bool   `json:"IsAMine"`
	IsRevealed       bool   `json:"IsRevealed"`
	IsFlagged        bool   `json:"IsFlagged"`
	FlaggedWith      string `json:"FlaggedWith"`
	MinesSurrounding int    `json:"MinesSurrounding"`
	Processed        bool
}

type CellGrid []Cell

type Game struct {
	GameID        string   `json:"GameID"`
	Rows          int      `json:"Rows"`
	Cols          int      `json:"Cols"`
	CantCells     int      `json:"CantCells"`
	RevealedCells int      `json:"RevealedCells"`
	Mines         int      `json:"Mines"`
	Clicks        int      `json:"Clicks"`
	Status        string   `json:"Status"`
	Board         CellGrid `json:"Board"`
}
