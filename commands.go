package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/w0/aggregator/internal/database"
)

type command struct {
	name      string
	arguments []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("The login handler expects a single argument, the username.\n")
	}

	username := cmd.arguments[0]

	_, err := s.db.GetUser(context.Background(), username)

	if err != nil {
		fmt.Printf("User: %s doesn't exist in the database.\n", username)
		os.Exit(1)
	}

	err = s.cfg.SetUser(username)

	if err != nil {
		return err
	}

	fmt.Printf("Set Username: %s\n", username)

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	return c.cmds[cmd.name](s, cmd)
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("The register handler expects a single argument, the username.\n")
	}

	now := time.Now()

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      cmd.arguments[0],
	})

	if err != nil {
		fmt.Printf("User: %s already exists in the database.\n", cmd.arguments[0])
		os.Exit(1)
	}

	err = s.cfg.SetUser(cmd.arguments[0])

	if err != nil {
		return err
	}

	fmt.Printf("User: %s has been created.\n", cmd.arguments[0])
	fmt.Printf("Data: %v\n", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())

	if err != nil {
		return err
	}

	fmt.Println("All users have been deleted from the database.")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())

	if err != nil {
		return err
	}

	for _, v := range users {
		out := fmt.Sprintf("* %s", v.Name)
		if v.Name == s.cfg.CurrentUserName {
			out = fmt.Sprintf("%s (current)", out)
		}
		fmt.Println(out)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("You must specify a url.")
	}

	feed, err := fetchFeed(context.Background(), cmd.arguments[0])

	if err != nil {
		return fmt.Errorf("Failed fetching rss feed: %w", err)
	}

	fmt.Printf("Data: %v", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("Usage: \"Name of Feed\" url..")
	}

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)

	if err != nil {
		return err
	}

	now := time.Now()

	feed, err := s.db.CreateFeed(context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Name:      cmd.arguments[0],
			Url:       cmd.arguments[1],
			UserID:    user.ID,
		})

	if err != nil {
		return err
	}

	fmt.Printf("Feed: %s has been added for user %s\n", feed.Name, user.Name)
	fmt.Printf("Data: %v", feed)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	res, err := s.db.GetFeeds(context.Background())

	if err != nil {
		return err
	}

	for _, v := range res {
		fmt.Printf("Feed: %s (%s), created by: %s\n", v.Name, v.Url, v.CreatedBy)
	}

	return nil
}
