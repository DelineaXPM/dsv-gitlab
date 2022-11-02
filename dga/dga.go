package dga

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	env "github.com/caarlos0/env/v6"
	"github.com/pterm/pterm"
)

// defaultTimeout defines default timeout for HTTP requests.
const defaultTimeout = time.Second * 5

// PermissionReadWriteOwner is the octal permission for Read Write for the owner of the file.
const PermissionReadWriteOwner = 0o600

type Config struct {
	IsCI    bool `env:"GITLAB_CI"`      // IsCI determines if the system is detecting being in CI system. https://docs.gitlab.com/ee/ci/variables/#enable-debug-logging
	IsDebug bool `env:"CI_DEBUG_TRACE"` // IsDebug is based on gitlab flagging as debug/trace level.

	CIProjectDirectory string `env:"CI_PROJECT_DIR,notEmpty"` // CIProjectDirectory is populated by CI_PROJECT_DIR which provides the fully qualified path to the project. https://docs.gitlab.com/ee/ci/variables/
	CIJobName          string `env:"CI_JOB_NAME,notEmpty"`    // CIJobName is populated by CI_JOB_NAME which provides the fully qualified path to the project. https://docs.gitlab.com/ee/ci/variables/
	// DSV SPECIFIC ENV VARIABLES.

	DomainEnv       string `env:"DSV_DOMAIN,notEmpty"`                 // Tenant domain name (e.g. example.secretsvaultcloud.com).
	ClientIDEnv     string `env:"DSV_CLIENT_ID,notEmpty"`              // Client ID for authentication.
	ClientSecretEnv string `json:"-" env:"DSV_CLIENT_SECRET,notEmpty"` // Client Secret for authentication.
	RetrieveEnv     string `env:"DSV_RETRIEVE,notEmpty"`               // JSON formatted string with data to retrieve from DSV.
}

// SecretToRetrieve defines JSON format of elements that expected in DSV_RETRIEVE list.
//nolint:tagliatelle // Here 'camel' casing is used instead of 'kebab'.
type SecretToRetrieve struct {
	SecretPath     string `json:"secretPath"`
	SecretKey      string `json:"secretKey"`
	OutputVariable string `json:"outputVariable"`
}

// getEnvFileName helps retrieve and build a env file path that should contain
// the resulting secrets. See [GitLab - Passing An Environment Variable to Another Job](https://docs.gitlab.com/ee/ci/variables/#pass-an-environment-variable-to-another-job)
func (cfg *Config) getEnvFileName() string {
	envFileName := filepath.Join(cfg.CIProjectDirectory, cfg.CIJobName)
	pterm.Debug.Printfln("envfilename: %s", envFileName)
	pterm.Success.Printfln("getEnvFileName() success")
	return envFileName
}

// configure Pterm settings for project based on the detected environment.
func (cfg *Config) configureLogging() {
	pterm.Info.Println("configureLogging()")

	pterm.Error = *pterm.Error.WithShowLineNumber().WithLineNumberOffset(1) //nolint:reassign // changing prefix later, not an issue.
	pterm.Warning = *pterm.Warning.WithShowLineNumber().WithLineNumberOffset(1)
	pterm.Warning = *pterm.Error.WithShowLineNumber().WithLineNumberOffset(1)
	pterm.Success.Printfln("configureLogging() success")
}

func (cfg *Config) sendRequest(c HTTPClient, req *http.Request, out any) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Delinea-DSV-Client", "gitlab-action")
	resp, err := c.Do(req)
	if err != nil {
		pterm.Error.Printfln("sendRequest: %+v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: %s", req.Method, req.URL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		pterm.Error.Printfln("sendRequest() unable to read response body: %+v", err)
		return fmt.Errorf("could not read response body: %w", err)
	}

	if err = json.Unmarshal(body, &out); err != nil {
		pterm.Error.Printfln("Unmarshal(): %+v", err)
		return fmt.Errorf("could not unmarshal response body: %w", err)
	}
	pterm.Success.Printfln("sendRequest() success")
	return nil
}

func parseConfig() (Config, error) {
	cfg := Config{}
	cfg.configureLogging()
	if err := env.Parse(&cfg, env.Options{
		// Prefix: "DSV_",.
	}); err != nil {
		pterm.Error.Printfln("env.Parse() %+v", err)
		return Config{}, fmt.Errorf("unable to parse env vars: %w", err)
	}
	pterm.Success.Println("parsed environment variables")
	return cfg, nil
}

func Run() error { //nolint:funlen,cyclop // funlen: this could use refactoring in future to break it apart more, but leaving as is at this time.
	var err error
	var retrievedValues []SecretToRetrieve

	cfg, err := parseConfig()
	if err != nil {
		return err
	}

	if cfg.IsDebug {
		pterm.Info.Println("DEBUG detected, setting debug output to enabled")
		pterm.EnableDebugMessages()
		pterm.Debug.Println("debug messages have been enabled")

		// No %v to avoid exposing secret values.
		pterm.Debug.Printfln("IsCI            : %v", cfg.IsCI)
		pterm.Debug.Printfln("IsDebug         : %v", cfg.IsDebug)

		pterm.Debug.Printfln("DomainEnv       : %v", cfg.DomainEnv)
		pterm.Debug.Println("ClientIDEnv     : ** value exists, but not exposing in logs **")
		pterm.Debug.Println("ClientSecretEnv : ** value exists, but not exposing in logs **")
		pterm.Debug.Printfln("RetrieveEnv     : %v", cfg.RetrieveEnv)
	}

	retrievedValues, err = ParseRetrieve(cfg.RetrieveEnv)
	if err != nil {
		pterm.Error.Printfln("run failure: %v", err)
		return err
	}

	apiEndpoint := fmt.Sprintf("https://%s/v1", cfg.DomainEnv)
	httpClient := &http.Client{Timeout: defaultTimeout}

	token, err := DSVGetToken(httpClient, apiEndpoint, &cfg)
	if err != nil {
		pterm.Error.Printfln("authentication failure: %v", err)
		return fmt.Errorf("unable to get access token")
	}
	var envFile *os.File

	// This function will only run if is CI.
	if cfg.IsCI {
		envFile, err = OpenEnvFile(&cfg)
		if err != nil {
			pterm.Error.Printfln("unable to run OpenEnvFile: %v", err)
			return err
		}
		defer envFile.Close()
	}

	for _, item := range retrievedValues {
		pterm.Debug.Printfln("start processing: SecretPath: %s SecretKey: %s", item.SecretPath, item.SecretKey)
		secret, err := DSVGetSecret(httpClient, apiEndpoint, token, item, &cfg)
		if err != nil {
			pterm.Error.Printfln("%q: Failed to fetch secret: %v", item, err)
			return fmt.Errorf("unable to get secret")
		}

		secretData, ok := secret["data"].(map[string]interface{})
		if !ok {
			pterm.Error.Printfln("%q: Cannot get data from secret", item)
			return fmt.Errorf("cannot parse secret")
		}
		pterm.Success.Printfln("retrieved successfully: %q", item)

		val, ok := secretData[item.SecretKey].(string)
		if !ok {
			pterm.Error.Printfln("%q: Key %q not found in data", item, item.SecretKey)
			return fmt.Errorf("specified field was not found in data")
		}

		pterm.Debug.Printfln("%q: Found %q key in data", item, item.SecretKey)

		if !cfg.IsCI {
			continue
		}

		outputKey := item.OutputVariable

		if err := ExportEnvVariable(envFile, outputKey, val); err != nil {
			pterm.Error.Printfln("%q: unable to export env variable: %v", outputKey, err)
			return fmt.Errorf("cannot set environment variable")
		}
		pterm.Success.Printfln("%q: Set env var %q to value in %q", item, strings.ToUpper(outputKey), item.SecretKey)
	}
	return nil
}

func ParseRetrieve(retrieve string) ([]SecretToRetrieve, error) {
	pterm.Info.Println("parseRetrieve()")

	var retrieveThese []SecretToRetrieve
	if err := json.Unmarshal([]byte(retrieve), &retrieveThese); err != nil {
		return []SecretToRetrieve{}, fmt.Errorf("unable to unmarshal: %w", err)
	}
	pterm.Success.Printfln("parseRetrieve(): returning %+v", retrieveThese)
	return retrieveThese, nil
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func DSVGetToken(c HTTPClient, apiEndpoint string, cfg *Config) (string, error) {
	pterm.Info.Println("DSVGetToken()")
	body := []byte(fmt.Sprintf(
		`{"grant_type":"client_credentials","client_id":"%s","client_secret":"%s"}`,
		cfg.ClientIDEnv, cfg.ClientSecretEnv,
	))
	endpoint := apiEndpoint + "/token"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	resp := make(map[string]interface{})
	if err = cfg.sendRequest(c, req, &resp); err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	token, ok := resp["accessToken"].(string)
	if !ok {
		return "", fmt.Errorf("could not read access token from response")
	}
	return token, nil
}

func DSVGetSecret(client HTTPClient, apiEndpoint, accessToken string, item SecretToRetrieve, cfg *Config) (map[string]interface{}, error) {
	pterm.Info.Println("dsvGetSecret()")
	// Endpoint := apiEndpoint + "/secrets/" + secretPath.
	endpoint, err := url.JoinPath(apiEndpoint, "secrets", item.SecretPath)
	if err != nil {
		pterm.Debug.Println("dsvGetSecret() problem with building url")
		return nil, fmt.Errorf("unable to build url: %w", err)
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		pterm.Debug.Printfln("dsvGetSecret(): endpoint: %q", endpoint)
		return nil, fmt.Errorf("could not build request: %w", err)
	}

	req.Header.Set("Authorization", accessToken)

	resp := make(map[string]interface{})
	if err = cfg.sendRequest(client, req, &resp); err != nil {
		pterm.Debug.Printfln("cfg.sendRequest() failure on sending request endpoint:%q req:%+v", endpoint, req)

		return nil, fmt.Errorf("API call failed: %w", err)
	}
	pterm.Success.Printfln("dsvGetSecret() success")
	return resp, nil
}

// OpenEnvFile storing secrets that can extend to another job or task in Gitlab.
// See [GitLab - Passing An Environment Variable to Another Job](https://docs.gitlab.com/ee/ci/variables/#pass-an-environment-variable-to-another-job)
func OpenEnvFile(cfg *Config) (envFile *os.File, err error) {
	pterm.Info.Println("OpenEnvFile()")
	envFileName := cfg.getEnvFileName()
	envFile, err = os.OpenFile(envFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, PermissionReadWriteOwner) //nolint:nosnakecase // these are standard package values and ok to leave snakecase.
	if errors.Is(err, os.ErrNotExist) {
		// See if we can provide some useful info on the existing permissions.
		return nil, fmt.Errorf("envfile doesn't exist or has denied permission %s: %w", envFileName, err)
	}
	if err != nil {
		return nil, fmt.Errorf("general error cannot open file %s: %w", envFileName, err)
	}
	pterm.Success.Printfln("OpenEnvFile() success")
	return envFile, nil
}

func ExportEnvVariable(envFile *os.File, key, val string) error {
	pterm.Info.Println("ExportEnvVariable()")
	if _, err := envFile.WriteString(fmt.Sprintf("%s=%s\n", strings.ToUpper(key), val)); err != nil {
		return fmt.Errorf("could not update %s environment file: %w", envFile.Name(), err)
	}
	pterm.Success.Printfln("ExportEnvVariable() success")
	return nil
}
