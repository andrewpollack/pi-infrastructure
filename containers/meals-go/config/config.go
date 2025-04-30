package config

import (
	"log"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Cfg is the global parsed config struct
var Cfg Config

// k is the internal koanf instance
var k = koanf.New(".")

type Config struct {
	App struct {
		Aisles    []string `koanf:"aisles"`
		PdfLayout []int    `koanf:"pdf_layout"`
	} `koanf:"app"`

	Server struct {
		AllowedOrigins []string `koanf:"allowed_origins"`
	} `koanf:"server"`

	AWS struct {
		Bucket struct {
			Name string `koanf:"name"`
			Key  string `koanf:"key"`
		} `koanf:"bucket"`
	} `koanf:"aws"`

	Email struct {
		Sender    string   `koanf:"sender"`
		Receivers []string `koanf:"receivers"`
	} `koanf:"email"`

	Database struct {
		Postgres struct {
			URL string `koanf:"url"`
		} `koanf:"postgres"`
	} `koanf:"database"`
}

// Load reads the config from file into the Cfg struct
func Load(path string) {
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if err := k.Unmarshal("", &Cfg); err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
	}
}
