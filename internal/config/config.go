package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const defaultHTTPAddress = ":9000"
const defaultFirestoreCollection = "github_stats_snapshots"

type Config struct {
	GitHubToken          string
	GitHubUsername       string
	HTTPAddress          string
	GoogleCloudProjectID string
	FirestoreCollection  string
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

	username := strings.TrimSpace(
		os.Getenv("GITHUB_USERNAME"),
	)

	if username == "" {
		return Config{}, errors.New(
			"GITHUB_USERNAME is required",
		)
	}

	token := strings.TrimSpace(
		os.Getenv("GITHUB_TOKEN"),
	)

	address := strings.TrimSpace(
		os.Getenv("HTTP_ADDRESS"),
	)
	if address == "" {
		address = defaultHTTPAddress
	}

	googleCloudProject := strings.TrimSpace(
		os.Getenv("GOOGLE_CLOUD_PROJECT"),
	)
	if googleCloudProject == "" {
		return Config{}, errors.New(
			"GOOGLE_CLOUD_PROJECT is required",
		)
	}

	firestoreCollection := strings.TrimSpace(
		os.Getenv("FIRESTORE_COLLECTION"),
	)
	if firestoreCollection == "" {
		firestoreCollection = defaultFirestoreCollection
	}

	return Config{
		GitHubToken:          token,
		GitHubUsername:       username,
		HTTPAddress:          address,
		GoogleCloudProjectID: googleCloudProject,
		FirestoreCollection:  firestoreCollection,
	}, nil
}
