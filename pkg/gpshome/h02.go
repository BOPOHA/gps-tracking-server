package gpshome

// https://stackoverflow.com/questions/30652661/gps-tracking-grtq
// https://dl.dropboxusercontent.com/s/azmaae3znl7gdgj/GPS%2BTracker%2BPlatform%2BCommunication%2BProtocol-From%2BWinnie%2BHuaSunTeK-V1.0.5-2017.pdf

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/nenadvasic/gps-tracking-server/internal/gpsserver"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	protocolName = "gpshome"
	comma        = `,`      // ,
	MsgStartByte = byte(42) // *
	MsgEndByte   = byte(35) // #
)

type PacketField int

const (
	IHDR PacketField = iota
	DevID
	CMD
	HHmmss
	S
	Latitude
	D
	Longitude
	G
	Speed
	Direction
	DDMMYY
	DevStatus
	pw
	count
	mcc
	mnc
)

func (p PacketField) String() string {
	return [...]string{
		"IHDR",
		"DevID",
		"CMD",
		"HHmmss",
		"S",
		"Latitude",
		"D",
		"Longitude",
		"G",
		"Speed",
		"Direction",
		"DDMMYY",
		"DevStatus",
		"pw",
		"count",
		"mcc",
		"mnc",
	}[p]
}

type GpsHomeProtocol struct {
	CmdReader chan []byte
}

func (p *GpsHomeProtocol) Handle(readbuff []byte, conn net.Conn) gpsserver.HandlerResponse {

	var res gpsserver.HandlerResponse
	var records []gpsserver.GpsRecord

	if len(readbuff) == 0 {
		log.Println("empty message")
		return res
	}
	record, err1 := p.parseMsg(readbuff)
	if err1 != nil {
		res.Error = err1
	} else {
		select {
		case cmd, ok := <-p.CmdReader:
			if ok {
				log.Printf("Value %s was read.\n", cmd)
				_, err2 := conn.Write(cmd)
				if err2 != nil {
					res.Error = err2
				}
			} else {
				log.Println("Channel closed!")
			}
		default:
			log.Println("No value ready, moving on.")
		}
	}
	res.Records = append(records, record)

	return res
}

func (p *GpsHomeProtocol) parseMsg(readbuff []byte) (record gpsserver.GpsRecord, err error) {
	log.Printf("Got message: %s.\n", string(readbuff))
	if readbuff[0] != MsgStartByte {
		return record, errors.New(fmt.Sprintf("invalid first byte, got message: %v", readbuff[:32]))
	}
	MsgEndIndex := bytes.IndexByte(readbuff, MsgEndByte)
	if MsgEndIndex == -1 {
		return record, errors.New(fmt.Sprintf("not found end msg marker, got message: %v", readbuff))
	}
	byteMsg := string(readbuff[1:MsgEndIndex])
	slicedMsg := strings.Split(byteMsg, comma)
	if len(slicedMsg) == 1 {
		decodedMsg, err := hex.DecodeString(slicedMsg[0])
		if err != nil {
			return record, errors.New(fmt.Sprintf("not enough words, got message: %v", readbuff))
		}
		p.CmdReader <- decodedMsg
		return record, errors.New(fmt.Sprintf("Got CMD message: %s", decodedMsg))
	}
	log.Printf("LOG sliceMsg: %+v", slicedMsg)
	record.Imei = slicedMsg[DevID]
	if rDate, err := time.Parse("020106", slicedMsg[DDMMYY]); err == nil {
		if rTime, err := time.Parse("150405", slicedMsg[HHmmss]); err == nil {
			record.GpsTime = rTime.AddDate(rDate.Year(), int(rDate.Month()), rDate.Day()).Unix()
			record.GpsTime = rDate.Add(
				time.Hour*time.Duration(rTime.Hour()) +
					time.Minute*time.Duration(rTime.Minute()) +
					time.Second*time.Duration(rTime.Second())).Unix()
		} else {
			println(err.Error())
		}
	} else {
		println(err.Error())
	}

	if glng, err := strconv.ParseFloat(slicedMsg[Longitude], 64); err == nil {
		if glat, err := strconv.ParseFloat(slicedMsg[Latitude], 64); err == nil {
			nLat := float64(int64(glat/100)) + (glat-float64(int64(glat/100)*100))/60
			nLng := float64(int64(glng/100)) + (glng-float64(int64(glng/100)*100))/60
			record.Location = gpsserver.GeoJson{
				Type:        "Point",
				Coordinates: []float64{nLng, nLat},
			}
		}
	}
	if speed, err := strconv.ParseFloat(slicedMsg[Speed], 32); err == nil {
		record.Speed = int(speed)
	}
	if direction, err := strconv.ParseFloat(slicedMsg[Direction], 32); err == nil {
		record.Course = float32(direction)
	}
	record.Timestamp = time.Now().Unix()
	record.Protocol = protocolName
	record.Valid = true
	log.Printf("LOG record: %+v", record)
	return record, nil
}
