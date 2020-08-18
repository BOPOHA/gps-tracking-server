package main

import (
	"encoding/json"
	"github.com/nenadvasic/gps-tracking-server/internal/gps_server"
	"github.com/nenadvasic/gps-tracking-server/pkg/gpshome"
	"github.com/nenadvasic/gps-tracking-server/pkg/ruptela"
	"github.com/nenadvasic/gps-tracking-server/pkg/teltonika"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	Host         string                   `json:"host"`
	Db           *gps_server.DbConfig     `json:"db"`
	GpsProtocols []gps_server.GpsProtocol `json:"protocols"`
}

func main() {

	log.Println("INFO", "Program pokrenut")

	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("ERROR", err)
	}

	config := Config{}

	err1 := json.NewDecoder(file).Decode(&config)
	if err1 != nil {
		log.Fatalln("ERROR", err1)
	}

	servers := gps_server.NewGpsServers(config.Db)
	log.Println("INFO", "Broj protokola u konfiguraciji:", len(config.GpsProtocols))

	for _, gpsProtocol := range config.GpsProtocols {

		if gpsProtocol.Enabled {

			var protocol_handler gps_server.GpsProtocolHandler

			switch gpsProtocol.Name {
			case "ruptela":
				protocol_handler = gps_server.GpsProtocolHandler(&ruptela.RuptelaProtocol{})
			case "teltonika":
				protocol_handler = gps_server.GpsProtocolHandler(&teltonika.TeltonikaProtocol{})
			case "gpshome":
				protocol_handler = gps_server.GpsProtocolHandler(
					&gpshome.GpsHomeProtocol{
						make(chan []byte, 3),
					},
				)
			default:
				log.Fatalln("ERROR", "Protocol handler nije definisan:", gpsProtocol.Name)
			}

			s := servers.NewGpsServer(gpsProtocol.Name, protocol_handler)

			s.Start(config.Host, gpsProtocol.Port)

			log.Println("INFO", "Server pokrenut za protokol "+gpsProtocol.Name+" na portu "+gpsProtocol.Port)
		}
	}

	log.Println("INFO", "Svi serveri su pokrenuti")

	// os.Exit(0);

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	// <-ch
	log.Println("INFO", "Dobijen signal", <-ch)

	servers.Stop()
	log.Println("INFO", "Program zaustavljen")
}
