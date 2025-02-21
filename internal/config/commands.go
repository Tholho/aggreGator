package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tholho/aggreGator/internal/database"
)

type State struct {
	CfgPtr *Config
	Db     *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	handlers map[string]func(*State, Command) error
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	c.handlers[name] = f
}

func (c *Commands) RegisterAll() {
	c.handlers = make(map[string]func(*State, Command) error)
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", middlewareLoggedIn(handlerReset))
	c.register("users", handlerUsers)
	c.register("agg", middlewareLoggedIn(handlerAgg))
	c.register("addfeed", middlewareLoggedIn(handlerAddfeed))
	c.register("feeds", middlewareLoggedIn(handlerFeeds))
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("error: unknown command '%s'", cmd.Name)
	}
	return handler(s, cmd)
}

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		currentUserRecord, err := s.Db.GetUser(context.Background(), s.CfgPtr.Current_user_name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("User not registered")
				return err
			}
			return err
		}
		return handler(s, cmd, currentUserRecord)
	}
}

func handlerBrowse(s *State, cmd Command, user database.User) error {
	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  2,
	}
	if len(cmd.Args) >= 1 {
		val, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		params.Limit = int32(val)
	}
	posts, err := s.Db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Println(post.Title)
		fmt.Println(post.Description)
	}
	return nil
}

func handlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Please enter a valid URL")
		return fmt.Errorf("invalid arguments")
	}
	params := database.DeleteFeedFollowParams{}
	params.Url = cmd.Args[0]
	params.UserID = user.ID
	err := s.Db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	return nil
}

func handlerFollowing(s *State, cmd Command, user database.User) error {
	name := user.Name
	res, err := s.Db.GetFeedFollowsForUser(context.Background(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("user follows nothing ?")
			return err
		}
		return err
	}
	fmt.Println(res)
	return nil
}

func handlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		fmt.Println("Please enter a valid url to follow")
		return fmt.Errorf("invalid arguments")
	}
	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.Args[0])
	fmt.Println(cmd.Args[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("unknown feed")
			return err
		}
		return err
	}
	params := database.CreateFeedFollowParams{}
	params.ID = uuid.New()
	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()
	params.FeedID = feed.ID
	params.UserID = user.ID
	_, er := s.Db.CreateFeedFollow(context.Background(), params)
	//fmt.Println(res)
	if er != nil {
		return er
	}
	fmt.Println(feed.Feedname, user.Name)
	return nil
}

func handlerFeeds(s *State, cmd Command, user database.User) error {
	//result := database.GetFeedsRow{}
	result, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		fmt.Println("--- debug ---\nError fetching feeds from db")
		return err
	}
	for _, row := range result {
		fmt.Println(row.Feedname, "-", row.Url, "-", row.Username)
	}
	return nil
}

func handlerAddfeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		fmt.Println("Please enter a name for the feed followed by the corresponding url")
		return fmt.Errorf("not enough arguments for command %s", cmd.Name)
	}
	createFeedParams := database.CreateFeedParams{}
	createFeedParams.ID = uuid.New()
	createFeedParams.CreatedAt = time.Now()
	createFeedParams.UpdatedAt = time.Now()
	createFeedParams.Name = cmd.Args[0]
	createFeedParams.Url = cmd.Args[1]
	createFeedParams.UserID = user.ID
	feedCreated, err := s.Db.CreateFeed(context.Background(), createFeedParams)
	if err != nil {
		fmt.Println("Could not create feed with parameters:\n", createFeedParams)
	}
	//fmt.Println("--- debug: new feed ---\n", feedCreated)
	params := database.CreateFeedFollowParams{}
	params.ID = uuid.New()
	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()
	params.FeedID = feedCreated.ID
	params.UserID = user.ID
	_, er := s.Db.CreateFeedFollow(context.Background(), params)
	if er != nil {
		return er
	}
	return nil
}

func handlerAgg(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		fmt.Println("please enter a duration like 1s, 1m, 1h, etc")
		return fmt.Errorf("invalid argument")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(timeBetweenRequests)
	fmt.Println("Collecting feeds every", timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Login command requires an argument")
		return fmt.Errorf("login command requires an argument")
	}
	s.CfgPtr.SetUser(cmd.Args[0])
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Register command requires an argument")
		return fmt.Errorf("register command requires an argument")
	}
	newUser, err := s.CreateUser(cmd.Args[0])
	if err != nil {
		return err
	}
	cmd.Name = "login"
	cmd.Args[0] = newUser.Name
	handlerLogin(s, cmd)
	return nil
}

func (s *State) CreateUser(name string) (database.User, error) {
	createUserParams := database.CreateUserParams{}
	createUserParams.ID = uuid.New()
	createUserParams.Name = name
	createUserParams.CreatedAt = time.Now()
	createUserParams.UpdatedAt = time.Now()
	userCreate, err := s.Db.CreateUser(context.Background(), createUserParams)
	if err != nil {
		return database.User{}, err
	}
	return userCreate, nil
}

func handlerUsers(s *State, cmd Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.CfgPtr.Current_user_name {
			fmt.Println("*", user.Name, "(current)")
		} else {
			fmt.Println("*", user.Name)
		}
	}
	return nil
}

func handlerReset(s *State, cmd Command, user database.User) error {
	err := s.Db.DeleteAllUsers(context.Background())
	if err != nil {
		fmt.Println("Did not fulfil reset request")
		return err
	}
	fmt.Println("Users database was erased entirely")
	return nil
}
