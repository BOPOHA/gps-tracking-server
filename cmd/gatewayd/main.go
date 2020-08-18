package main

import (
	"github.com/nenadvasic/gps-tracking-server/internal/gpsconfig"
	"github.com/nenadvasic/gps-tracking-server/internal/gpsserver"
	"github.com/nenadvasic/gps-tracking-server/pkg/gpshome"
	"github.com/nenadvasic/gps-tracking-server/pkg/ruptela"
	"github.com/nenadvasic/gps-tracking-server/pkg/teltonika"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	log.Println("INFO", "Program pokrenut")
	config := gpsconfig.ReadBaseConfigJson("")
	servers := gpsserver.NewGpsServers(config.Db)
	log.Println("INFO", "Broj protokola u konfiguraciji:", len(config.GpsProtocols))

	for _, gpsProtocol := range config.GpsProtocols {

		if gpsProtocol.Enabled {

			var protocolHandler gpsserver.GpsProtocolHandler

			switch gpsProtocol.Name {
			case "ruptela":
				protocolHandler = gpsserver.GpsProtocolHandler(&ruptela.RuptelaProtocol{})
			case "teltonika":
				protocolHandler = gpsserver.GpsProtocolHandler(&teltonika.TeltonikaProtocol{})
			case "gpshome":
				protocolHandler = gpsserver.GpsProtocolHandler(
					&gpshome.GpsHomeProtocol{
						CmdReader: make(chan []byte, 3),
					},
				)
			default:
				log.Fatalln("ERROR", "Protocol handler nije definisan:", gpsProtocol.Name)
			}

			s := servers.NewGpsServer(gpsProtocol.Name, protocolHandler, config.Db)

			s.Start(config.Host, gpsProtocol.Port)

			log.Println("INFO", "Server pokrenut za protokol "+gpsProtocol.Name+" na portu "+gpsProtocol.Port)
		}
	}

	log.Println("INFO", "Svi serveri su pokrenuti")

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	// <-ch
	log.Println("INFO", "Dobijen signal", <-ch)

	servers.Stop()
	log.Println("INFO", "Program zaustavljen")
}
