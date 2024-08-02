package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

var numbers = [75]int{}

func Randomizer() {
	for i := 0; i < 75; i++ {
		numbers[i] = i + 1
	}

	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
}

func init() {
	Randomizer()

	db, _ = gorm.Open(sqlite.Open("./database/bingo.db"), &gorm.Config{})

	db.AutoMigrate(&Match{})
}

type Match struct {
	ID         string    `json:"id"`
	Numbers    []int     `json:"numbers" gorm:"type:json"`
	IsOpen     bool      `json:"is_open"`
	LastNumber int       `json:"last_number"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MatchesRequest struct {
	Status  int     `json:"status"`
	Matches []Match `json:"matches"`
}

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, Match{
			uuid.New().String(),
			numbers[:],
			true,
			numbers[0],
			time.Now(),
			time.Now(),
		})
	})

	r.GET("/matches", func(c *gin.Context) {
		var matches []Match
		db.Find(&matches)

		c.JSON(http.StatusOK, MatchesRequest{
			http.StatusOK,
			matches,
		})
	})

	r.Run(":3000")
}
