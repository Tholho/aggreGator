package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("error: unknown command '%s'", cmd.Name)
	}

	return handler(s, cmd)
}

func handlerAgg(s *State, cmd Command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Println(feed)
	return nil
}

func handlerLogin(s *State, cmd Command) error {
	//fmt.Println(cmd.Args)
	if len(cmd.Args) < 1 {
		fmt.Println("Login command requires an argument")
		return fmt.Errorf("login command requires an argument")
	}
	userGet, err := s.Db.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("User not found")
			return err
		}
		return err
	}
	fmt.Println("THIS IS", userGet)
	s.CfgPtr.SetUser(cmd.Args[0])
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Register command requires an argument")
		return fmt.Errorf("register command requires an argument")
	}
	err := s.CreateUser(cmd.Args[0])
	if err != nil {
		return err
	}
	cmd.Name = "login"
	handlerLogin(s, cmd)
	return nil
}

func (s *State) CreateUser(name string) error {
	/*	userGet, err := s.Db.GetUser(context.Background(), name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("User not found")
				return err
			}
			return err
		}
		fmt.Println("THIS IS", userGet)
	*/
	createUserParams := database.CreateUserParams{}
	createUserParams.ID = uuid.New()
	createUserParams.Name = name
	createUserParams.CreatedAt = time.Now()
	createUserParams.UpdatedAt = time.Now()
	userCreate, err := s.Db.CreateUser(context.Background(), createUserParams)
	if err != nil {
		return err
	}
	fmt.Println("THIS IS", userCreate)
	return nil
	//(id, created_at, updated_at, name)
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

func handlerReset(s *State, cmd Command) error {
	err := s.Db.DeleteAllUsers(context.Background())
	if err != nil {
		fmt.Println("Did not fulfil reset request")
		return err
	}
	fmt.Println("Users database was erased entirely")
	return nil
}
