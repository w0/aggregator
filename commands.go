package main

import (
	"context"
	"fmt"
	"strconv"
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

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	return c.cmds[cmd.name](s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("the login handler expects a single argument, the username")
	}

	username := cmd.arguments[0]

	_, err := s.db.GetUser(context.Background(), username)

	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	err = s.cfg.SetUser(username)

	if err != nil {
		return err
	}

	fmt.Printf("Set Username: %s\n", username)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("the register handler expects a single argument, the username")
	}

	now := time.Now()

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      cmd.arguments[0],
	})

	if err != nil {
		return fmt.Errorf("user already exists: %w", err)
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
		return fmt.Errorf("specify a a duration 1s, 1m, 1h")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])

	if err != nil {
		return fmt.Errorf("failed to parse duration. %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	// ticker controls request loop. Loop each time the specified duration is reached.
	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			fmt.Printf("save erro: %v\n", err)
		}
	}

}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("usage: \"Name of Feed\" url")
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

	_, err = s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			FeedID:    feed.ID,
			UserID:    user.ID,
		})

	if err != nil {
		return fmt.Errorf("failed creating follow %w", err)
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("you must specify a url to follow")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.arguments[0])

	if err != nil {
		return err
	}

	now := time.Now()

	follow, err := s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			FeedID:    feed.ID,
			UserID:    user.ID,
		})

	if err != nil {
		return err
	}

	fmt.Printf("%s followed %s", follow.UserName, follow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("%s no follows. %w", user.Name, err)
	}

	fmt.Printf("%s is following:\n", user.Name)
	for _, v := range feeds {
		fmt.Printf("\t * %s\n", v.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("please specify the feed url to unfollow")
	}

	err := s.db.DeleteFeedFollow(context.Background(),
		database.DeleteFeedFollowParams{
			ID:  user.ID,
			Url: cmd.arguments[0],
		})

	if err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	fmt.Printf("unfollowed: %s", cmd.arguments[0])

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {

	limit, err := func(args []string) (int, error) {
		if len(args) > 0 {
			return strconv.Atoi(args[0])
		}
		return 2, nil

	}(cmd.arguments)

	if err != nil {
		return fmt.Errorf("failed to parse limit %w", err)
	}

	posts, err := s.db.GetPostsForUser(context.Background(),
		database.GetPostsForUserParams{
			UserID: user.ID,
			Limit:  int32(limit),
		})

	if err != nil {
		return fmt.Errorf("failed to fetch posts for user %w", err)
	}

	fmt.Printf("\n%s's latest posts\n\n", user.Name)
	for _, v := range posts {
		fmt.Printf("%s from %s\n", v.PublishedAt.Time.Format("Mon Jan 2"), v.FeedName)
		fmt.Printf("--- %s ---\n", v.Title)
		fmt.Printf("    %v\n", v.Description.String)
		fmt.Printf("Link: %s\n", v.Url)
		fmt.Println("=====================================")
	}

	return nil
}
