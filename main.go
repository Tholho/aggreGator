package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/tholho/aggreGator/internal/config"
	"github.com/tholho/aggreGator/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments. Please enter a command name")
		os.Exit(1)
	} else {
		cmd := config.Command{
			Name: os.Args[1],
			Args: os.Args[2:],
		}
		/*
			if len(os.Args) > 2 {
				cmd.Args = os.Args
			} else {
				cmd.Args = make([]string, 0)
			}
		*/
		cfg, err := config.Read()
		state := config.State{}
		state.CfgPtr = &cfg
		if err != nil {
			fmt.Println(err)
		} else {
			db, err := sql.Open("postgres", state.CfgPtr.Db_url)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer db.Close()
			dbQueries := database.New(db)
			state.Db = dbQueries
			fmt.Println(cmd)
			cmds := config.Commands{}
			cmds.RegisterAll()
			err = cmds.Run(&state, cmd)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			//	fmt.Println(cfg.Db_url)
			/*
				cfg.DisplayConfig()
				cfg.SetUser("bob")
				cfg, err = config.Read()
				if err != nil {
					fmt.Println(err)
				}
				cfg.DisplayConfig()
			*/
		}

	}
}
