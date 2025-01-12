package main

import (
	"fmt"

	"github.com/tholho/aggreGator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	} else {
		//	fmt.Println(cfg.Db_url)
		cfg.DisplayConfig()
		cfg.SetUser("bob")
		cfg, err = config.Read()
		if err != nil {
			fmt.Println(err)
		}
		cfg.DisplayConfig()

	}
}
