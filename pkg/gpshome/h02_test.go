package gpshome

import (
	"github.com/nenadvasic/gps-tracking-server/internal/gpsserver"
	"github.com/nenadvasic/gps-tracking-server/internal/tools"
	"testing"
)

const msg1 = "*HQ,7028114082,V1,145801,A,4217.4213,N,01850.4384,E,025.36,344,180820,FFFFBBFF,0,0,0,0#"

var (
	gpsRecord1 = gpsserver.GpsRecord{
		Imei: "7028114082",
		Location: gpsserver.GeoJson{
			Type:        "Point",
			Coordinates: []float64{18.84064, 42.290355},
		},
		Course:   344,
		Speed:    25,
		GpsTime:  1597762681,
		Protocol: protocolName,
		Valid:    true,
	}
	msgMatrix0 = []struct {
		msg string
		gps gpsserver.GpsRecord
	}{
		{msg1, gpsRecord1},
	}
)

func TestGpsHomeProtocol_Handle(t *testing.T) {
	h := GpsHomeProtocol{}
	for i, matrix := range msgMatrix0 {
		record, err := h.parseMsg([]byte(matrix.msg))
		if err != nil {
			t.Errorf("TestGpsHomeProtocol_Handle: %v err: %v\n", i, err)
		}
		record.Timestamp = 0
		tools.AssertEqual(t, matrix.gps, record)
	}

}
