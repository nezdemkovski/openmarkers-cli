package secrets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openmarkers/openmarkers-cli/internal/shared/constants"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const (
	KeyAccessToken  = "access-token"
	KeyRefreshToken = "refresh-token"
	KeyClientID     = "client-id"
	KeyClientSecret = "client-secret"

	Service = constants.AppName
)

type Source string

const (
	SourceEnv     Source = "environment"
	SourceKeyring Source = "keyring"
	SourceFile    Source = "config_file"
	SourceNone    Source = "none"
)

type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return e.Message
}

func Get(key string) (value string, source Source, err error) {
	envKey := toEnvVar(key)
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue, SourceEnv, nil
	}

	value, err = getFromKeyringWithTimeout(key)
	if err == nil && value != "" {
		return value, SourceKeyring, nil
	}
	keyringErr := err

	value, err = getFromFile(key)
	if err == nil && value != "" {
		return value, SourceFile, nil
	}

	if keyringErr != nil && !errors.Is(keyringErr, keyring.ErrNotFound) {
		return "", SourceNone, fmt.Errorf("failed to get secret %q (keyring error: %w)", key, keyringErr)
	}
	return "", SourceNone, nil
}

func Set(key, value string) (source Source, insecure bool, err error) {
	err = setInKeyringWithTimeout(key, value)
	if err == nil {
		_ = deleteFromFile(key)
		return SourceKeyring, false, nil
	}

	if fileErr := setInFile(key, value); fileErr != nil {
		return SourceNone, true, fmt.Errorf("failed to store secret in keyring (%w) and file (%w)", err, fileErr)
	}

	return SourceFile, true, nil
}

func Delete(key string) {
	_ = deleteFromKeyringWithTimeout(key)
	_ = deleteFromFile(key)
}

func DeleteAll() {
	for _, key := range []string{KeyAccessToken, KeyRefreshToken, KeyClientID, KeyClientSecret} {
		Delete(key)
	}
}

func toEnvVar(key string) string {
	envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
	return strings.ToUpper(constants.AppName) + "_" + envKey
}

func getFromKeyringWithTimeout(key string) (string, error) {
	ch := make(chan struct {
		val string
		err error
	}, 1)
	go func() {
		defer close(ch)
		val, err := keyring.Get(Service, key)
		ch <- struct {
			val string
			err error
		}{val, err}
	}()
	select {
	case res := <-ch:
		return res.val, res.err
	case <-time.After(3 * time.Second):
		return "", &TimeoutError{"timeout while trying to get secret from keyring"}
	}
}

func setInKeyringWithTimeout(key, value string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Set(Service, key, value)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return &TimeoutError{"timeout while trying to set secret in keyring"}
	}
}

func deleteFromKeyringWithTimeout(key string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Delete(Service, key)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return &TimeoutError{"timeout while trying to delete secret from keyring"}
	}
}

type secretsFile struct {
	Secrets map[string]string `yaml:"secrets"`
}

func secretsFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "."+constants.ConfigDirName, "secrets.yml")
}

func loadSecretsFile() (*secretsFile, error) {
	data, err := os.ReadFile(secretsFilePath())
	if err != nil {
		return &secretsFile{Secrets: make(map[string]string)}, err
	}
	var sf secretsFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return &secretsFile{Secrets: make(map[string]string)}, err
	}
	if sf.Secrets == nil {
		sf.Secrets = make(map[string]string)
	}
	return &sf, nil
}

func saveSecretsFile(sf *secretsFile) error {
	path := secretsFilePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(sf)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func getFromFile(key string) (string, error) {
	sf, err := loadSecretsFile()
	if err != nil {
		return "", err
	}
	return sf.Secrets[key], nil
}

func setInFile(key, value string) error {
	sf, _ := loadSecretsFile()
	sf.Secrets[key] = value
	return saveSecretsFile(sf)
}

func deleteFromFile(key string) error {
	sf, err := loadSecretsFile()
	if err != nil {
		return err
	}
	delete(sf.Secrets, key)
	if len(sf.Secrets) == 0 {
		return os.Remove(secretsFilePath())
	}
	return saveSecretsFile(sf)
}
