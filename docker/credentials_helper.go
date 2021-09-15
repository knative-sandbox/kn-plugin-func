package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/types"
	"github.com/docker/docker-credential-helpers/client"
)

var ErrCredentialsNotFound = errors.New("credentials not found")

func GetCredentialsFromCredsStore(registry string) (types.DockerAuthConfig, error) {
	result := types.DockerAuthConfig{}

	dirname, err := os.UserHomeDir()
	if err != nil {
		return result, fmt.Errorf("failed to determine home directory: %w", err)
	}

	confFilePath := filepath.Join(dirname, ".docker", "config.json")

	f, err := os.Open(confFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return types.DockerAuthConfig{}, ErrCredentialsNotFound
		}
		return result, fmt.Errorf("failed to open docker config file: %w", err)
	}
	defer f.Close()

	conf := struct {
		Store string `json:"credsStore"`
	}{}

	decoder := json.NewDecoder(f)

	err = decoder.Decode(&conf)
	if err != nil {
		return result, fmt.Errorf("failed to deserialize docker config file: %w", err)
	}

	if conf.Store == "" {
		return result, fmt.Errorf("no store configured")
	}

	helperName := fmt.Sprintf("docker-credential-%s", conf.Store)
	p := client.NewShellProgramFunc(helperName)

	credentialsMap, err := client.List(p)
	if err != nil {
		return result, fmt.Errorf("failed to list credentials: %w", err)
	}

	for serverUrl := range credentialsMap {
		if to2ndLevelDomain(serverUrl) == to2ndLevelDomain(registry) {
			creds, err := client.Get(p, serverUrl)
			if err != nil {
				return result, fmt.Errorf("failed to get credentials: %w", err)
			}
			result.Username = creds.Username
			result.Password = creds.Secret
			return result, nil
		}
	}

	return result, fmt.Errorf("failed to get credentials from helper specified in ~/.docker/config.json: %w", ErrCredentialsNotFound)
}

func to2ndLevelDomain(rawurl string) string {
	if !strings.Contains(rawurl, "://") {
		rawurl = "https://" + rawurl
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return ""
	}
	hostname := u.Hostname()
	parts := strings.Split(hostname, ".")
	if len(parts) <= 1 {
		return hostname
	}
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}
