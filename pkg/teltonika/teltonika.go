/**
 * Teltonika Protocol
 */
package teltonika

import (
	"bytes"
	"encoding/binary"

	"errors"
	"github.com/nenadvasic/gps-tracking-server/internal/gpsserver"
	"github.com/nenadvasic/gps-tracking-server/internal/tools"
	"log"
	"net"
	"time"
)

const (
	TeltonikaProtocolName = "teltonika"
	TeltonikaCodecGh3000  = 0x07
	TeltonikaCodecFm4x00  = 0x08
	TeltonikaCodec12      = 0x0C
)

type TeltonikaProtocol struct {
}

func (p *TeltonikaProtocol) Handle(readbuff []byte, conn net.Conn) gpsserver.HandlerResponse {
	buff := bytes.NewBuffer(readbuff)

	var startBytes uint16
	var imei string

	binary.Read(buff, binary.BigEndian, &startBytes)

	res := gpsserver.HandlerResponse{}

	// Ako imamo nešto u prva 2 bajta onda je uređaj poslao svoj IMEI
	if startBytes > 0 {

		res.Imei = p.getIMEI(buff)

		log.Println("INFO", "Device IMEI:", res.Imei)

		_, err := conn.Write([]byte{0x01}) // ACK
		if err != nil {
			res.Error = err
		}
		// Ako su prva dva bajta nule onda je uređaj poslao GPS zapise
	} else {

		res.Imei = imei

		records, err1 := p.getRecords(buff, imei)
		if err1 != nil {
			res.Error = err1
		}
		res.Records = records

		// Šaljemo ACK
		err2 := binary.Write(conn, binary.BigEndian, int32(len(records)))
		if err2 != nil {
			res.Error = err2
		}
	}

	return res
}

func (p *TeltonikaProtocol) getIMEI(buff *bytes.Buffer) string {

	var imei string

	buff.Truncate(15)

	imei = buff.String()

	if imei == "" {
		// TODO ?
	}

	return tools.PadLeft(imei, "0", 15)
}

func (p *TeltonikaProtocol) getRecords(buff *bytes.Buffer, imei string) ([]gpsserver.GpsRecord, error) {

	var records []gpsserver.GpsRecord

	var dataLength uint32
	var codec byte
	var priority byte     // ne koristimo za sada
	var recordsCount byte // broj recorda u tekućem zahtevu
	var gpstime uint64
	var lon int32
	var lat int32
	var alt int16
	var course uint16
	var sat byte
	var speed uint16

	buff.Next(2)

	binary.Read(buff, binary.BigEndian, &dataLength)
	binary.Read(buff, binary.BigEndian, &codec)

	if codec == TeltonikaCodec12 {
		// TODO ?
		return nil, errors.New("CODEC 12")
	}

	binary.Read(buff, binary.BigEndian, &recordsCount)

	log.Println("INFO", "Broj recorda u zahtevu:", recordsCount)

	for i := 0; i < int(recordsCount); i++ {

		if codec == TeltonikaCodecGh3000 {
			// TODO
		} else {
			binary.Read(buff, binary.BigEndian, &gpstime)
			binary.Read(buff, binary.BigEndian, &priority)

			binary.Read(buff, binary.BigEndian, &lon)
			binary.Read(buff, binary.BigEndian, &lat)
			binary.Read(buff, binary.BigEndian, &alt)
			binary.Read(buff, binary.BigEndian, &course)
			binary.Read(buff, binary.BigEndian, &sat)
			binary.Read(buff, binary.BigEndian, &speed)

			lonFloat := float64(lon) / 10000000
			latFloat := float64(lat) / 10000000

			if !tools.IsValidCoordinates(latFloat, lonFloat) {
				log.Println("ERROR", "Nepravilne vrednosti koordinata! IMEI:", imei, "Lon:", lonFloat, "Lat:", latFloat)
				continue
			}

			location := gpsserver.GeoJson{Type: "Point", Coordinates: []float64{lonFloat, latFloat}}
			sensors := make([]gpsserver.GpsSensor, 0) // TODO

			isValid := tools.IsValidRecord(sat)

			record := gpsserver.GpsRecord{Imei: imei, Location: location, Altitude: float32(alt) / 10, Course: float32(course) / 100, Speed: int(speed), Satellites: int(sat), Sensors: sensors, GpsTime: int64(gpstime / 1000), Timestamp: time.Now().Unix(), Protocol: TeltonikaProtocolName, Valid: isValid}

			// log.Println(record)

			records = append(records, record)

			buff.Next(6) // TODO senzori
		}
	}

	return records, nil
}
