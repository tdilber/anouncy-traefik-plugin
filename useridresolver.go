package anouncy_traefik_plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	LogDebug = log.New(ioutil.Discard, "DEBUG: user-id-resolver: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogInfo  = log.New(ioutil.Discard, "INFO: user-id-resolver: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogWarn  = log.New(ioutil.Discard, "WARN: user-id-resolver: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// Config the plugin configuration.
type Config struct {
	resolverUrl string `yaml:"resolverurl"`
	logLevel    string `yaml:"loglevel"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type UserResolveResult struct {
	userId          string `json:"userId"`
	anonymousUserId string `json:"anonymousUserId"`
}

// New creates and returns a plugin instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	LogWarn.SetOutput(os.Stdout)
	LogWarn.Println("config", config)
	if config.logLevel == "DEBUG" {
		LogDebug.SetOutput(os.Stdout)
		LogInfo.SetOutput(os.Stdout)
		LogWarn.SetOutput(os.Stdout)
	} else if config.logLevel == "INFO" {
		LogInfo.SetOutput(os.Stdout)
		LogWarn.SetOutput(os.Stdout)
	} else if config.logLevel == "WARN" {
		LogWarn.SetOutput(os.Stdout)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		LogDebug.Println("BEGIN")

		authHeader, ok := req.Header["Authorization"]
		if ok {
			LogDebug.Println("authHeader", authHeader)
			requestURL := fmt.Sprintf("%s%s", config.resolverUrl, authHeader[0])
			res, err := http.Get(requestURL)
			if err != nil {
				errorMessage := fmt.Sprintf("error making http request: %s\n", err)
				LogWarn.Println(errorMessage)
			} else {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					errorMessage := fmt.Sprintf("client: could not read response body: %s\n", err)
					LogWarn.Println(errorMessage)
				} else {
					result := UserResolveResult{}
					err := json.Unmarshal(resBody, &result)
					LogDebug.Println("request json result:", string(resBody[:]))

					if err != nil {
						errorMessage := fmt.Sprintf("client: json parse: %s\n", err)
						LogWarn.Println(errorMessage)
					} else {
						LogDebug.Println("USER-ID", result.userId)
						req.Header.Set("USER-ID", result.userId)
						LogDebug.Println("ANONYMOUS-USER-ID", result.anonymousUserId)
						req.Header.Set("ANONYMOUS-USER-ID", result.anonymousUserId)
					}
				}
			}
		} else {
			LogDebug.Println("Header Not Found!")
		}

		LogDebug.Println("END")
		next.ServeHTTP(rw, req)
	}), nil
}
