/**
 * GPS Tracking Server
 */
package gps_server

import (
	"bufio"
	"gopkg.in/mgo.v2"
	"log"
	"net"
	"sync"
	"time"
)

type GpsProtocolHandler interface {
	Handle([]byte, net.Conn) HandlerResponse
}

type HandlerResponse struct {
	Error   error
	Imei    string
	Records []GpsRecord
}

type GpsProtocol struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Port    string `json:"port"`
	Enabled bool   `json:"enabled"`
}

type GeoJson struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// TODO
type GpsSensor struct {
	SensorId string
}

// type GpsDevice struct {
// 	Imei      string
// 	IpAddress string
// 	// Active bool
// }

type DbConfig struct {
	Host string `json:"host"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Name string `json:"name"`
	Col  string `json:"col"`
}

type GpsRecord struct {
	Imei     string  `json:"imei"`
	Location GeoJson `json:"location"`
	//         Latitude    float64             `json:"lat"`
	//         Longitude   float64             `json:"lon"`
	Altitude   float32     `json:"alt"`
	Course     float32     `json:"course"`
	Speed      int         `json:"speed"`
	Satellites int         `json:"satellites"`
	Sensors    []GpsSensor `json:"sensors"`
	GpsTime    int         `json:"gpstime"` // vreme dobijeno od uređaja
	Timestamp  int         `json:"timestamp"`
	Protocol   string      `json:"protocol"`
	Valid      bool        `json:"valid"` // Zapis smatramo validnim ako ima 3+ satelita
}

type GpsServer struct {
	name         string
	mu           *sync.RWMutex
	mongoSession *mgo.Session
	listener     net.Listener
	protocol     GpsProtocolHandler
}

type GpsServers struct {
	mongoSession *mgo.Session
	servers      map[string]*GpsServer
}

func NewGpsServers(dbConfig *DbConfig) *GpsServers {

	mongoSession, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{dbConfig.Host},
		Username: dbConfig.User,
		Password: dbConfig.Pass,
		Database: dbConfig.Name,
	})
	// sessionCopy.SetMode(mgo.Monotonic, true)

	if err != nil {
		log.Fatalln("ERROR", "Neuspešno konektovanje na bazu:", err)
	}
	return &GpsServers{
		mongoSession: mongoSession,
		servers:      make(map[string]*GpsServer),
	}
}

func (ss *GpsServers) NewGpsServer(name string, protocol GpsProtocolHandler) *GpsServer {

	log.Println("INFO", "Inicijalizacija servera:", name)

	s := &GpsServer{
		name:         name,
		mongoSession: ss.mongoSession,
		protocol:     protocol,
	}

	return s
}

func (s *GpsServer) Start(host string, port string) {

	log.Println("INFO", "Pokretanje servera:", s.name) // + " on [" + host + ":" + port + "] ...")
	s.Listen(host, port)
	go s.Serve()
}

func (s *GpsServer) Listen(host string, port string) {

	listener, err := net.Listen("tcp4", host+":"+port)
	if err != nil {
		log.Fatalln("ERROR", "Program nije u mogućnosti da otvori listening socket:", err.Error())
	}

	log.Println("INFO", "Socket uspešno otvoren na", listener.Addr())

	s.listener = listener
}

func (s *GpsServer) Serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			panic(err)
		}
		go s.HandleRequest(conn)
	}
}

func (s *GpsServer) HandleRequest(conn net.Conn) {

	defer log.Println("INFO", "Disconnecting:", conn.RemoteAddr())
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		readbuff := scanner.Bytes()
		res := s.protocol.Handle(readbuff, conn)
		if res.Error != nil {
			log.Println("ERROR", res.Error)
			return
		}

		if len(res.Records) > 0 {
			s.SaveGpsRecords(res.Records)
		}
	}
}

func (s *GpsServer) SaveGpsRecords(records []GpsRecord) bool {

	sessionCopy := s.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB("dbname").C("collection")

	for _, record := range records {
		err1 := c.Insert(record)
		if err1 != nil {
			log.Println("ERROR", "Neuspešan upis recorda u bazu:", err1)
			return false
		}

		log.Println("INFO", "Record sačuvan", record.Imei, record.Location.Coordinates, record.Speed, record.Sensors, time.Unix(int64(record.GpsTime), 0), record.Protocol)
	}

	return true
}

func (ss *GpsServers) Stop() {

	for _, s := range ss.servers {
		s.listener.Close()
	}
}
