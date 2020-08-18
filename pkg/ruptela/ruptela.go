/**
 * Ruptela Protocol
 */
package ruptela

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/nenadvasic/gps-tracking-server/internal/gpsserver"
	"github.com/nenadvasic/gps-tracking-server/internal/tools"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	RuptelaCommandRecords = 0x01
	RuptelaProtocolName   = "ruptela"
)

type RuptelaProtocol struct {
}

func (p *RuptelaProtocol) Handle(readbuff []byte, conn net.Conn) gpsserver.HandlerResponse {

	res := gpsserver.HandlerResponse{}

	buff := bytes.NewBuffer(readbuff)

	records, err1 := p.getRecords(buff)
	if err1 != nil {
		res.Error = err1
	}
	res.Records = records

	// Šaljemo ACK
	_, err2 := conn.Write([]byte{0x00, 0x02, 0x64, 0x01, 0x13, 0xbc})
	if err2 != nil {
		res.Error = err2
	}

	return res
}

func (p *RuptelaProtocol) getRecords(buff *bytes.Buffer) ([]gpsserver.GpsRecord, error) {

	var records []gpsserver.GpsRecord

	var imei uint64
	var tip byte          // tip zahteva
	var recordsLeft byte  // broj preostalih recorda na uređaju (ne koristimo za sada)
	var recordsCount byte // broj recorda u tekućem zahtevu
	var gpstime uint32
	var lon int32
	var lat int32
	var alt uint16
	var course uint16
	var sat byte
	var speed uint16

	// buff := bytes.NewBuffer(readbuff)

	// log.Printf("%x", buff)

	buff.Next(2)

	binary.Read(buff, binary.BigEndian, &imei)
	binary.Read(buff, binary.BigEndian, &tip)

	imeiString := tools.PadLeft(strconv.FormatUint(imei, 10), "0", 15)

	// log.Println("INFO", "Device IMEI:", imeiString)

	if tip != RuptelaCommandRecords {
		log.Println("ERROR", "Nepoznat tip zahteva:", tip)
		return nil, errors.New("nepoznat tip zahteva")
	}

	binary.Read(buff, binary.BigEndian, &recordsLeft)
	binary.Read(buff, binary.BigEndian, &recordsCount)

	log.Println("INFO", "Broj recorda u zahtevu:", recordsCount)

	for i := 0; i < int(recordsCount); i++ {

		binary.Read(buff, binary.BigEndian, &gpstime)

		buff.Next(2)

		binary.Read(buff, binary.BigEndian, &lon)
		binary.Read(buff, binary.BigEndian, &lat)
		binary.Read(buff, binary.BigEndian, &alt)
		binary.Read(buff, binary.BigEndian, &course)
		binary.Read(buff, binary.BigEndian, &sat)
		binary.Read(buff, binary.BigEndian, &speed)

		lonFloat := float64(lon) / 10000000
		latFloat := float64(lat) / 10000000

		if !tools.IsValidCoordinates(latFloat, lonFloat) {
			log.Println("ERROR", "Nepravilne vrednosti koordinata! IMEI:", imeiString, "Lon:", lonFloat, "Lat:", latFloat)
			continue
		}

		location := gpsserver.GeoJson{Type: "Point", Coordinates: []float64{lonFloat, latFloat}}
		sensors := make([]gpsserver.GpsSensor, 0) // TODO

		buff.Next(2)

		// Senzori mogu da šalju podatke u setovima veličine 1/2/4/8 bajtova
		// Podaci su naslagani redom sa prefix bajtom koji predstavlja broj bajtova u setu (bytes_count)
		var bytesCount byte
		var sensorId byte
		var data1 byte
		var data2 uint16
		var data4 uint32
		var data8 uint64

		// Read 1 byte data
		binary.Read(buff, binary.BigEndian, &bytesCount)
		// fmt.Println(bytes_count)
		for j := 0; j < int(bytesCount); j++ {
			binary.Read(buff, binary.BigEndian, &sensorId)
			binary.Read(buff, binary.BigEndian, &data1)
			// TODO: Dodavanje u slice sensors
		}

		// Read 2 byte data
		binary.Read(buff, binary.BigEndian, &bytesCount)
		// fmt.Println(bytes_count)
		for j := 0; j < int(bytesCount); j++ {
			binary.Read(buff, binary.BigEndian, &sensorId)
			binary.Read(buff, binary.BigEndian, &data2)
			// TODO: Dodavanje u slice sensors
		}

		// Read 4 byte data
		binary.Read(buff, binary.BigEndian, &bytesCount)
		// fmt.Println(bytes_count)
		for j := 0; j < int(bytesCount); j++ {
			binary.Read(buff, binary.BigEndian, &sensorId)
			binary.Read(buff, binary.BigEndian, &data4)
			// TODO: Dodavanje u slice sensors
		}

		// Read 8 byte data
		binary.Read(buff, binary.BigEndian, &bytesCount)
		// fmt.Println(bytes_count)
		for j := 0; j < int(bytesCount); j++ {
			binary.Read(buff, binary.BigEndian, &sensorId)
			binary.Read(buff, binary.BigEndian, &data8)
			// TODO: Dodavanje u slice sensors
		}

		isValid := tools.IsValidRecord(sat)

		record := gpsserver.GpsRecord{Imei: imeiString, Location: location, Altitude: float32(alt) / 10, Course: float32(course) / 100, Speed: int(speed), Satellites: int(sat), Sensors: sensors, GpsTime: int64(gpstime), Timestamp: time.Now().Unix(), Protocol: RuptelaProtocolName, Valid: isValid}

		records = append(records, record)
	}

	return records, nil
}
