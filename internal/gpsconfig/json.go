package gpsconfig

import (
	"encoding/json"
	"github.com/nenadvasic/gps-tracking-server/internal/gpsserver"
	"log"
	"os"
)

type Config struct {
	Host         string                  `json:"host"`
	Db           *gpsserver.DbConfig     `json:"db"`
	GpsProtocols []gpsserver.GpsProtocol `json:"protocols"`
}

func ReadBaseConfigJson(path string) Config {
	var config Config

	if path == "" {
		path = "config.json"
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatalln("ERROR", err)
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatalln("ERROR", err)
	}

	return config

}
