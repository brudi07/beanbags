package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RoundRequest struct {
	MatchID  int `json:"match_id" binding:"required"`
	RoundNum int `json:"round_num"` // optional; can be auto-calculated
}

func (h *Handler) CreateRound(c *gin.Context) {
	var req RoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Automatically determine the next round number if not provided
	if req.RoundNum == 0 {
		row := h.DB.QueryRow("SELECT IFNULL(MAX(round_num), 0) + 1 FROM rounds WHERE match_id = ?", req.MatchID)
		if err := row.Scan(&req.RoundNum); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get next round number"})
			return
		}
	}

	// Create new round
	stmt, err := h.DB.Prepare(`
		INSERT INTO rounds (match_id, round_num, start_time)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare insert"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(req.MatchID, req.RoundNum, time.Now().UTC())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert round"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Round created successfully",
		"match_id":  req.MatchID,
		"round_num": req.RoundNum,
	})
}
