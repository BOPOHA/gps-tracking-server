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

	count_protocols := len(config.GpsProtocols)

	log.Println("INFO", "Broj protokola u konfiguraciji:", count_protocols)

	var servers []*gps_server.GpsServer

	host := config.Host

	for i := 0; i < count_protocols; i++ {

		protocol_name := config.GpsProtocols[i].Name
		protocol_port := config.GpsProtocols[i].Port

		if config.GpsProtocols[i].Enabled {

			var protocol_handler gps_server.GpsProtocolHandler
			switch protocol_name {
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
				log.Fatalln("ERROR", "Protocol handler nije definisan:", protocol_name)
			}

			s := gps_server.NewGpsServer(protocol_name, config.Db, protocol_handler)

			s.Start(host, protocol_port)

			// log.Println("INFO", "Server pokrenut za protokol " + protocol_name + " na portu " + protocol_port)

			servers = append(servers, s)
		}
	}

	log.Println("INFO", "Svi serveri su pokrenuti")

	// os.Exit(0);

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	// <-ch
	log.Println("INFO", "Dobijen signal", <-ch)

	stopServers(servers)

	// time.Sleep(10000 * time.Millisecond)
	log.Println("INFO", "Program zaustavljen")
}

func stopServers(servers []*gps_server.GpsServer) {

	for _, server := range servers {
		server.Stop()
	}
}
