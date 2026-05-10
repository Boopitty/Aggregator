package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Boopitty/Aggregator/internal/config"
	"github.com/Boopitty/Aggregator/internal/database"
	"github.com/google/uuid"
)

// The state is the user's current states,
// such as the current user, the current session, etc.
type state struct {
	cfg *config.Config
	db  *database.Queries
}

// The command struct represents a command that the user can run,
type command struct {
	name  string
	slice []string
}

// The commands struct holds a map of command names to their handler functions.
type commands struct {
	handlers map[string]func(*state, command) error
}

// Runs a given command with the provided state if it exists.
func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("Command not found: %s", cmd.name)
	}
	return handler(s, cmd) // Run the func and return the error value it produces.
}

// Registers a new handler function for a command name.
func (c *commands) register(name string, f func(*state, command) error) error {
	if f == nil {
		return fmt.Errorf("Handler function cannot be nil")
	}
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = f
	return nil
}

// Prints the details of a feed fetched from a URL.
func agg(s *state, cmd command) error {
	if len(cmd.slice) == 0 {
		return fmt.Errorf("No time duration provided: 5s, 1m, 1h, etc.")
	}

	// Get the time string and turn it into a time.Duration object
	time_between_reqs := cmd.slice[0]
	timer, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("Error parsing time duration: %w", err)
	}
	fmt.Printf("Collecting feeds every %s...\n", timer)

	// Every time the duration ends, call the scrapeFeeds() function
	ticker := time.NewTicker(timer)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		fmt.Println("Attemting Scrape:")
		err = scrapeFeeds(s)
		if err != nil {
			fmt.Printf("Error Scraping Feeds: %v\n", err)
			os.Exit(1)
		}
	}
}

// Scrapes the feeds from the database and saves the posts to the database.
func scrapeFeeds(s *state) error {
	// Get the next feed to fetch from the database
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting next feed to fetch: %w", err)
	}
	if feed.ID == uuid.Nil {
		fmt.Println("No feeds to fetch")
		return nil
	}

	// Mark the feed as fetched in the database
	err = s.db.MarkFeedFetch(context.Background(), database.MarkFeedFetchParams{
		ID:            feed.ID,
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("Error marking feed as fetched: %w", err)
	}
	fmt.Printf("Posting Feed: %s\n", feed.Name)

	// Fetch the feed using the URL from the database
	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("Error fetching feed data: %w", err)
	}

	currentTime := time.Now()

	// Create a post for the feed
	err = s.db.CreatePost(context.Background(), database.CreatePostParams{
		ID:          uuid.New(),
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
		Url:         feedData.Channel.Link,
		Title:       sql.NullString{String: feedData.Channel.Title, Valid: true},
		Description: sql.NullString{String: feedData.Channel.Description, Valid: true},
		FeedID:      feed.ID,
	})
	if err != nil {
		// if the Error is because of a duplicate key value, update that post.
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			post, err := s.db.GetPostByURL(context.Background(), feedData.Channel.Link)

			err = s.db.UpdatePost(context.Background(), database.UpdatePostParams{
				ID:          post.ID,
				UpdatedAt:   currentTime,
				Title:       sql.NullString{String: feedData.Channel.Title, Valid: true},
				Description: sql.NullString{String: feedData.Channel.Description, Valid: true},
			})
			if err != nil {
				return fmt.Errorf("Error Updating Post: %w", err)
			}

		} else {
			return fmt.Errorf("Error creating post: %w", err)
		}

	}

	// Create a new post in the database for each item in the feed
	for _, item := range feedData.Channel.Item {
		// Check if post with same URL exits. If it does, update it. Otherwise create one.
		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
			Title:       sql.NullString{String: item.Title, Valid: true},
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: sql.NullString{String: item.PubDate, Valid: true},
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				post, err := s.db.GetPostByURL(context.Background(), feedData.Channel.Link)
				err = s.db.UpdatePost(context.Background(), database.UpdatePostParams{
					ID:          post.ID,
					UpdatedAt:   currentTime,
					Title:       sql.NullString{String: item.Title, Valid: true},
					Description: sql.NullString{String: item.Description, Valid: true},
				})
				if err != nil {
					return fmt.Errorf("Error Updating Post: %w", err)
				}
			} else {
				return fmt.Errorf("Error updating post: %w", err)
			}
		}
	}

	fmt.Printf("Finished Posting.\n\n")
	return nil
}

// Middleware function that checks if a user is logged in before allowing access to certain command handlers.
func middlewareLoggedIn(handler func(*state, command, *database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return fmt.Errorf("No user logged in. Please log in to use this command.")
		}

		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Error getting user: %w", err)
		}

		return handler(s, cmd, &user)
	}
}

// Prints all posts followed by the current user.
func handlerBrowse(s *state, cmd command, user *database.User) error {
	// Determine the limit of the number of outputs. The default is 2.
	limit := int64(2)
	if len(cmd.slice) > 0 {
		var err error
		limit, err = strconv.ParseInt(cmd.slice[0], 10, 32)
		if err != nil {
			return fmt.Errorf("Error parsing limit: %w", err)
		}
	}

	// Get the feeds followed by the user from the database
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Error getting feed follows for user: %w", err)
	}
	if len(feeds) == 0 {
		fmt.Printf("No feeds followed by %s\n", user.Name)
		return nil
	}

	// For each followed feed, get all the posts
	for _, feed := range feeds {
		fmt.Printf("Posts for feed: %s\n\n", feed.FeedName)

		// Get all the posts for the feed
		posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
			FeedID: feed.FeedID,
			Limit:  int32(limit),
		})
		if err != nil {
			return fmt.Errorf("Error getting posts: %w", err)
		}

		// Print the details of each post
		for _, post := range posts {
			fmt.Printf("Title: %v\nDescription: %v\nURL: %v\n\n", post.Title, post.Description, post.Url)
		}

	}
	return nil
}

// This will be the function signature of all command handlers.
func handlerLogin(s *state, cmd command) error {
	if cmd.slice == nil {
		return fmt.Errorf("No name provided")
	}
	name := cmd.slice[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("User not found")
		os.Exit(1)
	}
	// Set the user to the fist element of the slice, which should be the username.
	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("Error setting user: %w", err)
	}

	fmt.Println("User set successfully")
	return nil
}

// Register a new user into the database and set the current user to the new user's name.
func handlerRegister(s *state, cmd command) error {
	// This is a placeholder for the register command handler.
	if cmd.slice == nil {
		return fmt.Errorf("No name provided")
	}

	name := cmd.slice[0]
	// Check if the user already exists in the database.
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		fmt.Println("User already exists")
		os.Exit(1)
	}

	// Create a new user in the database with the provided name.
	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("Error creating user: %w", err)
	}

	// Set the current user in the config to the new user's name.
	s.cfg.SetUser(newUser.Name)
	fmt.Printf(
		"User registered successfully:\nID: %s\nName: %s\nCreated At: %s\nUpdated At: %s\n",
		newUser.ID, newUser.Name, newUser.CreatedAt, newUser.UpdatedAt,
	)
	return nil
}

// Empty the contents of the database
func handlerReset(s *state, cmd command) error {
	// This is a placeholder for the reset command handler.
	err := s.db.Reset(context.Background())
	if err != nil {
		fmt.Println("Error resetting database: %w", err)
		os.Exit(1)
	}
	fmt.Println("Database reset successfully")
	return nil
}

// Get the names of all users and mark the current user
func handlerGetAll(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error getting users:", err)
		os.Exit(1)
	}
	for _, user := range users {
		if user == s.cfg.CurrentUserName {
			fmt.Printf("%s (current)\n", user)
		} else {
			fmt.Println(user)
		}
	}
	return nil
}

// Add a new feed to the Feeds table and print its details.
func handlerAddFeed(s *state, cmd command, user *database.User) error {
	if cmd.slice == nil {
		return fmt.Errorf("No arguments provided")
	}
	if len(cmd.slice) < 2 {
		return fmt.Errorf("Not enough arguments provided")
	}

	feedName := cmd.slice[0]
	url := cmd.slice[1]

	// Check if the feed already exists in the database.
	feed, err := s.db.SearchFeedURL(context.Background(), url)
	if err == nil {
		fmt.Printf("Feed already exists:\nID: %s\nName: %s\nURL: %s\nCreated At: %s\nUpdated At: %s\n",
			feed.ID, feed.Name, feed.Url, feed.CreatedAt, feed.UpdatedAt,
		)
		return nil
	}

	time := time.Now()

	newFeed, err := s.db.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time,
		UpdatedAt: time,
		Name:      feedName,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("Error adding feed: %w", err)
	}

	fmt.Printf("Feed added successfully:\nID: %s\nName: %s\nURL: %s\nCreated At: %s\nUpdated At: %s\n",
		newFeed.ID, newFeed.Name, newFeed.Url, newFeed.CreatedAt, newFeed.UpdatedAt,
	)

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time,
		UpdatedAt: time,
		UserID:    user.ID,
		FeedID:    newFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}

	fmt.Printf("Feed followed successfully:\nUser: %s\nFeed: %s\n", user.Name, newFeed.Name)
	return nil
}

// Print the details of all feeds in the feeds table.
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.Feeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting feeds: %w", err)
	}

	for i, feed := range feeds {
		user, err := s.db.SearchUserID(context.Background(), feeds[i].UserID)
		if err != nil {
			return fmt.Errorf("Error searching user ID: %w", err)
		}
		fmt.Printf("Name: %s\nURL: %s\nUser Name ID: %s\n\n",
			feed.Name, feed.Url, user.Name)
	}
	return nil
}

// Follow a feed by its URL and print the details of the follow.
func handlerFollow(s *state, cmd command, user *database.User) error {
	if cmd.slice == nil {
		return fmt.Errorf("No arguments provided")
	}

	feed, err := s.db.SearchFeedURL(context.Background(), cmd.slice[0])
	if err != nil {
		return fmt.Errorf("Error searching feed: %w", err)
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}

	fmt.Println("Feed followed successfully:")
	fmt.Printf("ID: %s\nCreated At: %s\nUpdated At: %s\nUser: %s\nFeed: %s\n",
		follow.ID, follow.CreatedAt, follow.UpdatedAt, follow.UserName, follow.FeedName)
	return nil
}

// Print the names of all feeds followed by the current user.
func handlerFollowing(s *state, cmd command, user *database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Error getting feed follows for user: %w", err)
	}

	fmt.Printf("Feeds followed by %s:\n", user.Name)
	for _, follow := range follows {
		fmt.Printf("- %s\n", follow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user *database.User) error {
	if cmd.slice == nil {
		return fmt.Errorf("No arguments provided")
	}
	url := cmd.slice[0]
	feed, err := s.db.SearchFeedURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Error searching feed: %w", err)
	}

	err = s.db.UnfollowFeed(context.Background(), database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error unfollowing feed: %w", err)
	}

	fmt.Printf("Feed unfollowed successfully:\nUser: %s\nFeed: %s\n", user.Name, feed.Name)
	return nil
}
