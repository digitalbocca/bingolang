package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	Randomizer()

	var err error
	db, err = gorm.Open(sqlite.Open("./database/bingo.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Match{})
}

type IntSlice []int

func (s IntSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *IntSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, s)
}

type Match struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Numbers    IntSlice  `json:"numbers" gorm:"type:json"`
	IsOpen     bool      `json:"is_open"`
	LastNumber int       `json:"last_number"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MatchSummary struct {
	ID     string `json:"id"`
	IsOpen bool   `json:"is_open"`
}

func Randomizer() []int {
	var numbers = [75]int{}

	for i := 0; i < 75; i++ {
		numbers[i] = i + 1
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	r.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	return numbers[:] // Converte o array para um slice
}

func main() {
	r := gin.Default()

	r.GET("/matches", func(c *gin.Context) {
		var matches []Match
		if err := db.Find(&matches).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var summaries []MatchSummary
		for _, match := range matches {
			summaries = append(summaries, MatchSummary{
				ID:     match.ID,
				IsOpen: match.IsOpen,
			})
		}

		c.JSON(http.StatusOK, summaries)
	})

	r.GET("/matches/:id", func(c *gin.Context) {
		id := c.Param("id")
		var match Match
		if err := db.Where("id = ?", id).First(&match).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
			return
		}

		c.JSON(http.StatusOK, match)
	})

	r.POST("/matches", func(c *gin.Context) {
		// Verificar se há alguma partida aberta
		var openMatch Match
		if err := db.Where("is_open = ?", true).First(&openMatch).Error; err == nil {
			// Fechar a partida aberta
			openMatch.IsOpen = false
			if err := db.Save(&openMatch).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Criar uma nova partida aberta
		match := Match{
			ID:         uuid.New().String(),
			Numbers:    Randomizer(),
			IsOpen:     true,
			LastNumber: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := db.Create(&match).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, match)
	})

	r.Run(":3000")
}
