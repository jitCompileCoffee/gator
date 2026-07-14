package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jitCompileCoffee/blog-agg/internal/config"
	"github.com/jitCompileCoffee/blog-agg/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.cmds[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

func handlerLogins(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the login handler expects a single argument: the username")
	}
	username := cmd.args[0]
	if _, err := s.db.GetUser(context.Background(), username); err != nil {
		return err
	}
	if err := s.cfg.SetUser(username); err != nil {
		return fmt.Errorf("could not set user: %w", err)
	}
	fmt.Printf("User has been set to: %s\n", username)
	return nil
}

func register(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the register handler expects a single argument: username")
	}
	username := cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      username,
	})
	if err != nil {
		return err
	}
	if err := s.cfg.SetUser(username); err != nil {
		return err
	}
	fmt.Printf("user: %+v was created", user)
	return nil
}

func getUsers(s *state, _ command) error {
	currentUser := s.cfg.CurrentUserName
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user == currentUser {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s \n", user)
		}
	}
	return nil
}

func reset(s *state, _ command) error {
	if err := s.db.DeleteAllUsers(context.Background()); err != nil {
		return err
	}
	fmt.Println("database reset successful")
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("usage: gator agg <time_between_reqs>")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}
	fmt.Printf("Collecting feeds every %v...\n", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Printf("Scraper error: %v\n", err)
		}
	}
}

func hanlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("the addFeed handler expects two arguments: name, url")
	}
	name := cmd.args[0]
	url := cmd.args[1]
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Added feed: %+v and now following: %+v", feed, feedFollow)
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("--- Feeds in Database ---")
	for _, feed := range feeds {
		fmt.Printf("* Name: %s\n", feed.Name)
		fmt.Printf("  URL:  %s\n", feed.Url)
		fmt.Printf("  By:   %s\n", feed.UserName)
		fmt.Println("-------------------------")
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("the follow handler expects one argument: url")
	}
	url := cmd.args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created a new feed follow: Feed Name: %s, User: %s\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerGetFollows(s *state, cmd command, user database.User) error {
	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("could not retrieve followed feeds: %w", err)
	}
	if len(followedFeeds) == 0 {
		fmt.Println("You are not following any feeds yet!")
		return nil
	}
	fmt.Printf("Feeds followed by %s:\n", user.Name)
	for _, follow := range followedFeeds {
		fmt.Printf("* %s\n", follow.FeedName)
	}
	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("the unfollow handler expects one argument: url")
	}
	url := cmd.args[0]
	err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    url,
	})
	if err != nil {
		return fmt.Errorf("could not unfollow feed: %w", err)
	}
	fmt.Printf("Successfully unfollowed feed at %s\n", url)
	return nil
}

func handleBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit value: %w (must be a number)", err)
		}
		if parsedLimit <= 0 {
			return fmt.Errorf("limit must be greater than 0")
		}
		limit = parsedLimit
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("could not retrieve posts: %w", err)
	}
	if len(posts) == 0 {
		fmt.Println("No posts found for your followed feed")
		return nil
	}
	fmt.Printf("\n--- Browsing Latest %d Posts for %s ---\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("\n★  %s\n", post.Title)
		fmt.Printf("   Link: %s\n", post.Url)
		if post.PublishedAt.Valid {
			fmt.Printf("   Published: %s\n", post.PublishedAt.Time.Local().Format("2006-01-02 15:04"))
		}
		if post.Description.Valid && post.Description.String != "" {
			fmt.Printf("   Description: %s\n", post.Description.String)
		}
		fmt.Println("--------------------------------------------------")
	}

	return nil
}
