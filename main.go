package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/w0/aggregator/internal/config"
	"github.com/w0/aggregator/internal/database"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	dbQueries := database.New(db)

	s := state{
		cfg: &cfg,
		db:  dbQueries,
	}

	cmds := commands{
		cmds: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)

	if len(os.Args) < 2 {
		fmt.Println("Must enter a command and the args.")
		return
	}

	cmd := command{
		name:      os.Args[1],
		arguments: os.Args[2:],
	}

	if _, ok := cmds.cmds[cmd.name]; !ok {
		log.Fatalf("Command: %s is not available.", cmd.name)
	}

	if cmd.name == "agg" {
		cmd.arguments = make([]string, 1)
		cmd.arguments[0] = "https://www.wagslane.dev/index.xml"
	}

	err = cmds.run(&s, cmd)

	if err != nil {
		log.Fatalf("Command %s failed. %v", os.Args[1], err)
	}
}
