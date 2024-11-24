package main

import (
	"github.com/w0/aggregator/internal/config"
	"github.com/w0/aggregator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}
