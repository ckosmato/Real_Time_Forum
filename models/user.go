package models

import (
	"time"
)


type User struct {
	ID            string
	Nickname      string
	Age		      int
	Gender        string
	FirstName     string
	LastName      string
	Email         string
	Password      string
	createdAt     time.Time
}

