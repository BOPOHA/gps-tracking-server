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
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	http.HandleFunc("/geojson/", locationHandler.getGPXxml)
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
		println("creating a new session for: " + r.RemoteAddr)
		l.NewMongoSession()
	}

	sessionCopy := l.MongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(l.DbConfig.Name).C(l.DbConfig.Col)

	var record gpsserver.GpsRecord
	var isFirstResponce = true

	for {
		for shortName, device := range l.Devices {

			err := c.Find(bson.M{"imei": device.Imei}).Sort("-gpstime").Limit(1).One(&record)

			if err != nil {
				continue
			}

			if record.GpsTime > device.Timestamp || isFirstResponce {

				coor := record.Location.Coordinates
				marker := MapMarker{shortName, record.Imei, record.GpsTime, coor[0], coor[1], record.Speed}
				out, err := json.Marshal(marker)
				if err != nil {
					log.Println("Marshaling MapMarker error:", err)
					continue
				}

				if err := conn.WriteMessage(websocket.TextMessage, out); err != nil {
					log.Println("err write msg to websocket:", err, string(out))
					return
				}

				device.Timestamp = record.GpsTime
				isFirstResponce = false
			} else {
				if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Println("closing websocket: ", r.RemoteAddr, err)
					return
				}
			}
		}
		time.Sleep(5 * time.Second)
	}

}

func (l LocationHTTPHandler) getGPXxml(w http.ResponseWriter, r *http.Request) {

	log.Println("Req: ", r.RemoteAddr, r.RequestURI)
	split := strings.Split(r.RequestURI, "/")
	if len(split) < 4 {
		http.Error(w, "not enough params", http.StatusBadRequest)
	}
	deviceImei := split[2]
	strMinTimestamp := split[3]
	strMaxTimestamp := split[4]

	if len(strMinTimestamp) < 13 || len(strMaxTimestamp) < 13 {
		http.Error(w, "bad timestamps", http.StatusBadRequest)
		return
	}
	minRange, err := strconv.ParseInt(strMinTimestamp[:10], 10, 64)
	if err != nil {
		http.Error(w, "cant parse minRange", http.StatusBadRequest)
		return
	}
	maxRange, err := strconv.ParseInt(strMaxTimestamp[:10], 10, 64)
	if err != nil {
		http.Error(w, "cant parse maxRange", http.StatusBadRequest)
		return
	}
	var records []gpsserver.GpsRecord

	if l.MongoSession == nil {
		log.Println("creating a new session: " + r.RemoteAddr)
		l.NewMongoSession()
	}

	sessionCopy := l.MongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(l.DbConfig.Name).C(l.DbConfig.Col)

	_ = c.Find(bson.M{
		"imei":    deviceImei,
		"gpstime": bson.M{"$gt": minRange, "$lt": maxRange},
	}).All(&records)

	points := make([]GpsJSONFeature, 0)

	for id, v := range records {
		point := GpsJSONFeature{
			Type:     "Feature",
			Geometry: v.Location,
			Properties: GpsJSONProperties{
				PointId:   id,
				Speed:     v.Speed,
				Course:    v.Course,
				GpsTime:   time.Unix(v.GpsTime, 0),
				Timestamp: v.Timestamp,
				Sat:       v.Satellites,
			},
		}
		points = append(points, point)
	}
	rest := GpsJSONFeatures{
		Type:     "FeatureCollection",
		Features: points,
	}

	v, err := json.Marshal(rest)
	if err != nil {
		log.Println("err marshaling geojson:", err)
	}
	w.Header().Add("Content-Type", "application/vnd.geo+json")
	w.WriteHeader(http.StatusOK)
	w.Write(v)

}

type GpsJSONProperties struct {
	PointId   int
	Speed     int
	Course    float32
	GpsTime   time.Time
	Timestamp int64
	Sat       int
}

type GpsJSONFeature struct {
	Type     string `json:"type"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties GpsJSONProperties `json:"properties,omitempty"`
}

type GpsJSONFeatures struct {
	Type     string           `json:"type"`
	Features []GpsJSONFeature `json:"features"`
}

//var points []geojson.Object
//for _, v := range records {
//	point := geojson.NewPoint(geometry.Point{
//		X: v.Location.Coordinates[1],
//		Y: v.Location.Coordinates[0],
//	})
//	point.AppendJSON([]byte("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZzz"))
//	points = append(points, point)
//}
//gjsonFCollection := geojson.NewFeatureCollection(points)
//w.Header().Add("Content-Type", "application/vnd.geo+json")
//w.WriteHeader(http.StatusOK)
//w.Write(gjsonFCollection.AppendJSON(nil))

//wptRecords := []*gpx.WptType{}
//
//for _, v := range records {
//	wptR := gpx.WptType{
//		Lat:           v.Location.Coordinates[1],
//		Lon:           v.Location.Coordinates[0],
//		Ele:           float64(v.Altitude),
//		Speed:         float64(v.Speed),
//		Course:        float64(v.Course),
//		Time:          time.Unix(v.Timestamp, 0),
//		MagVar:        0,
//		GeoidHeight:   0,
//		Name:          "",
//		Cmt:           "",
//		Desc:          "",
//		Src:           "",
//		Link:          nil,
//		Sym:           "",
//		Type:          v.Location.Type,
//		Fix:           "",
//		Sat:           v.Satellites,
//		HDOP:          0,
//		VDOP:          0,
//		PDOP:          0,
//		AgeOfDGPSData: 0,
//		DGPSID:        nil,
//		Extensions:    nil,
//	}
//
//	wptRecords = append(wptRecords, &wptR)
//}
//
//w.Write([]byte(xml.Header))
//
//g := &gpx.GPX{
//	Version: "1.0",
//	Creator: "GPS-TRACKER IMEI: " + deviceImei,
//	Wpt:     wptRecords,
//}
//
//if err := g.WriteIndent(w, "", "  "); err != nil {
//	fmt.Printf("err == %v", err)
//}
