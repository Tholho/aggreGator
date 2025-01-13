package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func (c *Config) DisplayConfig() {
	data, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println(string(data))
}

func Read() (Config, error) {
	cfgPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	cfgData := Config{}
	file, err := os.Open(cfgPath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfgData)
	if err != nil {
		return Config{}, err
	}
	return cfgData, nil
}

func (c *Config) SetUser(username string) error {
	if username == "" {
		return fmt.Errorf("please enter a username")
	}
	c.Current_user_name = username
	write(c)
	return nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cfgPath := homeDir + "/" + configFileName
	return cfgPath, nil
}

func write(cfg *Config) error {
	cfgPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(cfgPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Beautifier le JSON
	if err := encoder.Encode(cfg); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return err
	}
	return nil
}
