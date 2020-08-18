//go:generate go build -o /tmp/pkger github.com/markbates/pkger/cmd/pkger
//go:generate /tmp/pkger -o cmd/frontend/
package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/markbates/pkger"
	"github.com/nenadvasic/gps-tracking-server/internal/gpsconfig"
	"github.com/nenadvasic/gps-tracking-server/internal/gpsserver"
	"gopkg.in/mgo.v2"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Device struct {
	Imei      string
	Timestamp int64
}

type MapMarker struct {
	SName   string  `json:"name"`
	Imei    string  `json:"imei"`
	GpsTime int64   `json:"gpstime"`
	Lon     float64 `json:"lon"`
	Lat     float64 `json:"lat"`
	Speed   int     `json:"speed"`
}

type LocationHTTPHandler struct {
	MongoSession *mgo.Session
	DbConfig     *gpsserver.DbConfig
	Devices      map[string]*Device
}

func main() {
	box := pkger.Dir("/web")
	locationHandler := NewLocationHTTPHandler()
	http.Handle("/location/", locationHandler)
	http.Handle("/", http.FileServer(box))
	err := http.ListenAndServe(":1337", nil)
	if err != nil {
		panic("Error: " + err.Error())
	}
}
func NewLocationHTTPHandler() LocationHTTPHandler {

	devices := make(map[string]*Device)
	devices["356307043490167"] = &Device{"356307043490167", 0}
	devices["012896004329949"] = &Device{"012896004329949", 0}
	devices["013227002640237"] = &Device{"013227002640237", 0}
	devices["terik"] = &Device{"7028114082", 0}

	return LocationHTTPHandler{
		MongoSession: nil,
		DbConfig:     gpsconfig.ReadBaseConfigJson("").Db,
		Devices:      devices,
	}
}

func (l *LocationHTTPHandler) NewMongoSession() {

	mongoSession, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{l.DbConfig.Host},
		Username: l.DbConfig.User,
		Password: l.DbConfig.Pass,
		Database: l.DbConfig.Name,
	})
	// sessionCopy.SetMode(mgo.Monotonic, true)

	if err != nil {
		log.Fatalln("ERROR", "Neuspešno konektovanje na bazu:", err)
	}
	l.MongoSession = mongoSession
}

func (l LocationHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	if l.MongoSession == nil {
		println("creating a new session" + "dddd")
		l.NewMongoSession()
	}

	sessionCopy := l.MongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(l.DbConfig.Name).C(l.DbConfig.Col)

	var record gpsserver.GpsRecord

	for {
		for shortName, device := range l.Devices {

			err := c.Find(bson.M{"imei": device.Imei}).Sort("-gpstime").One(&record)

			if err != nil {
				continue
			}

			if record.GpsTime > device.Timestamp {

				coor := record.Location.Coordinates
				marker := MapMarker{shortName, record.Imei, record.GpsTime, coor[0], coor[1], record.Speed}
				out, err := json.Marshal(marker)
				if err != nil {
					log.Println("Marshaling MapMarker error:", err)
					continue
				}

				err1 := conn.WriteMessage(websocket.TextMessage, out)
				if err1 != nil {
					log.Println("err write msg to websocket:", err, string(out))
					return
				}

				device.Timestamp = record.GpsTime
			}
		}
		time.Sleep(5 * time.Second)
	}

}
