package config

import (
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	StorePath string
}

func Default() (Config, error) {
	dir, err := defaultDataDir()
	if err != nil {
		return Config{}, err
	}
	return Config{StorePath: filepath.Join(dir, "tasks.json")}, nil
}

func defaultDataDir() (string, error) {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "todo"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", errors.New("could not resolve home directory")
	}
	return filepath.Join(home, ".local", "share", "todo"), nil
}
