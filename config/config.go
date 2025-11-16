package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	defaultConfigFileName = ".filmeta.json"
	langFName             = "./languages.json"
)

type Language struct {
	Name        string `json:"name,omitempty"`
	EnglishName string `json:"english_name,omitempty"`
	ISO639_1    string `json:"iso_639_1,omitempty"`
}
type Config struct {
	InProduction bool   `json:"inProduction,omitempty"`
	AppRoot      string `json:"appRoot,omitempty"`
	HugoRoot     string `json:"hugoRoot,omitempty"`
	TMDB         struct {
		APIKey       string `json:"apiKey,omitempty"`
		APIBase      string `json:"apiBase,omitempty"`
		PosterBase   string `json:"posterBase,omitempty"`
		BackdropBase string `json:"backdropBase,omitempty"`
	} `json:"tmdb"`
	Db struct {
		User                 string `json:"user,omitempty"`
		Passwd               string `json:"passwd,omitempty"`
		Net                  string `json:"net,omitempty"`
		Addr                 string `json:"addr,omitempty"`
		DBName               string `json:"dbName,omitempty"`
		ParseTime            bool   `json:"parseTime,omitempty"`
		Loc                  string `json:"loc,omitempty"`
		AllowNativePasswords bool   `json:"allowNativePasswords,omitempty"`
	} `json:"db"`
	Security struct {
		CSRFKey string `json:"csrfKey,omitempty"`
	} `json:"security"`
	Session struct {
		Name              string `json:"name,omitempty"`
		Path              string `json:"path,omitempty"`
		Domain            string `json:"domain,omitempty"`
		MaxAgeHours       int    `json:"maxAgeHours,omitempty"`
		AuthenticationKey string `json:"authenticationKey,omitempty"`
		EncryptionKey     string `json:"encryptionKey,omitempty"`
	} `json:"session"`
	Algolia struct {
		AppID     string `json:"appID"`
		SearchKey string `json:"searchKey"`
		WriteKey  string `json:"writeKey"`
	} `json:"algolia"`
}

var c = Config{}
var iso2lang = make(map[string]string, 0)

func Configuration(configFileName ...string) (*Config, error) {

	if (c == Config{}) {

		var cfName string
		switch len(configFileName) {
		case 0:
			dirname, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			cfName = fmt.Sprintf("%s/%s", dirname, defaultConfigFileName)
		case 1:
			cfName = configFileName[0]
		default:
			return nil, fmt.Errorf("incorrect arguments for configuration file name")
		}

		viper.SetConfigFile(cfName)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}

		if err := viper.Unmarshal(&c); err != nil {
			return nil, err
		}
	}

	return &c, nil
}

func initLang() error {

	cfg, err := Configuration()
	if err != nil {
		return fmt.Errorf("error reading configuration: %w", err)
	}

	langF, err := os.Open(filepath.Join(cfg.AppRoot, "config", "languages.json"))
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", langFName, err)
	}
	jsonBytes, err := io.ReadAll(langF)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", langFName, err)
	}

	langData := make(map[string]Language, 0)
	if err := json.Unmarshal(jsonBytes, &langData); err != nil {
		return fmt.Errorf("error unmarshaling language: %w", err)
	}

	for key, value := range langData {
		iso2lang[key] = value.EnglishName
	}

	return nil
}

func ISOLanguage(iso2 string) (string, error) {

	if len(iso2lang) == 0 {
		if err := initLang(); err != nil {
			return "", fmt.Errorf("error initializing language map: %w", err)
		}
	}

	return iso2lang[iso2], nil
}
