package gpshome

// https://stackoverflow.com/questions/30652661/gps-tracking-grtq
// https://dl.dropboxusercontent.com/s/azmaae3znl7gdgj/GPS%2BTracker%2BPlatform%2BCommunication%2BProtocol-From%2BWinnie%2BHuaSunTeK-V1.0.5-2017.pdf

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/nenadvasic/gps-tracking-server/internal/gps_server"
	"log"
	"net"
	"strings"
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

func (p *GpsHomeProtocol) Handle(readbuff []byte, conn *net.TCPConn, _ string) gps_server.HandlerResponse {

	res := gps_server.HandlerResponse{}

	select {
	case cmd, ok := <-p.CmdReader:
		if ok {
			fmt.Printf("Value %s was read.\n", cmd)
			_, err2 := conn.Write(cmd)
			if err2 != nil {
				res.Error = err2
			}
		} else {
			fmt.Println("Channel closed!")
		}
	default:
		fmt.Println("No value ready, moving on.")
	}

	records, err1 := p.parseMsg(readbuff)
	if err1 != nil {
		res.Error = err1
	}
	res.Records = records

	return res
}

func (p *GpsHomeProtocol) parseMsg(readbuff []byte) ([]gps_server.GpsRecord, error) {
	if readbuff[0] != MsgStartByte {
		return nil, errors.New(fmt.Sprintf("invalid first byte, got message: %v", readbuff[:32]))
	}
	MsgEndIndex := bytes.IndexByte(readbuff, MsgEndByte)
	if MsgEndIndex == -1 {
		return nil, errors.New(fmt.Sprintf("not found end msg marker, got message: %v", readbuff))
	}
	byteMsg := string(readbuff[1:MsgEndIndex])
	slicedMsg := strings.Split(byteMsg, comma)
	if len(slicedMsg) == 1 {
		decodedMsg, err := hex.DecodeString(slicedMsg[0])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("not enough words, got message: %v", readbuff))
		}
		p.CmdReader <- decodedMsg
		return nil, errors.New(fmt.Sprintf("Got CMD message: %s", decodedMsg))
	}
	log.Printf("%+v", slicedMsg)

	var records []gps_server.GpsRecord

	//lon_float, _ := strconv.ParseFloat(Message[Longitude], 64)
	//lat_float, _ := strconv.ParseFloat(Message[Latitude], 64)

	//records := []gps_server.GpsRecord{
	//{
	//	Message[DevID],
	//	gps_server.GeoJson{"Point", []float64{lon_float, lat_float}},
	//	0,
	//	0,
	//	0,
	//	0,
	//	nil,
	//	0,
	//	int(time.Now().Unix()),
	//	protocolName,
	//	true,
	//},
	//}

	return records, nil
}
