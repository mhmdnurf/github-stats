package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const defaultHTTPAddress = ":9000"

type Config struct {
	GitHubToken string
	HTTPAddress string
}

func Load() (Config, error) {
	return load(".env")
}

func load(filename string) (Config, error) {
	if err := godotenv.Load(filename); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf(
			"load %s: %w",
			filename,
			err,
		)
	}

	token := strings.TrimSpace(
		os.Getenv("GITHUB_TOKEN"),
	)
	if token == "" {
		return Config{}, errors.New(
			"GITHUB_TOKEN is required",
		)
	}

	address := strings.TrimSpace(
		os.Getenv("HTTP_ADDRESS"),
	)
	if address == "" {
		address = defaultHTTPAddress
	}

	return Config{
		GitHubToken: token,
		HTTPAddress: address,
	}, nil
}
