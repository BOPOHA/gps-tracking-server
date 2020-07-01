package gpshome

import (
	"encoding/hex"
	"fmt"
	"testing"
)

// 77.222.24.97 - - [29/Jun/2020 22:32:20] "*HQ,7028114082,V1,203218,A,4217.3750,N,01850.6526,E,000.00,281,290620,FFFFBBFF,0,0,0,0#$p(@ ) B5B
// 2020/06/29 23:43:04 INFO Record sačuvan 5849110696187932977 [82.550456 84.1761068] 13111 [] 1996-09-03 18:00:49 +0200 CEST gpshome
// 2020/06/29 23:39:48 INFO Record sačuvan 5849110696187932977 [82.5767216 77.53045] 12336 [] 1998-10-25 14:49:02 +0100 CET gpshome

const msg1 = "2a48512c373032383131343038322c56312c3231343934322c412c343231372e333736332c4e2c30313835302e363534332c" +
	"452c3030302e30302c3030302c3239303632302c46464646424246462c302c302c302c302300000000000000000000000000"

var (
	badInMsg = [][]byte{
		[]byte{
			36, 112, 40, 17, 64, 130, 35, 0, 5, 5, 7, 32, 66, 23, 56, 22, 65, 1, 133, 6, 128, 60, 0, 0, 0, 255, 255, 187, 255, 0, 0, 0, 0, 1, 41, 2, 79, 109, 40, 214, 13, 79, 109, 40, 216, 9, 79, 109, 39, 135, 8, 8, 0, 0,
		},
	}
)

func TestGpsHomeProtocol_Handle(t *testing.T) {
	h := GpsHomeProtocol{}
	decoded, err := hex.DecodeString(msg1)
	//fmt.Printf("Decoded: %#v\n", decoded)
	_, err = h.parseMsg(decoded)

	fmt.Printf("Len: %v, err: %v\n", len(decoded), err)
	if err != nil {
		t.Errorf("Len: %v, err: %v\n", len(decoded), err)
	}
}

func TestGpsHomeProtocol_Handle2(t *testing.T) {
	h := GpsHomeProtocol{}
	for _, msg := range badInMsg {
		//fmt.Printf("Decoded: %#v\n", decoded)
		_, err := h.parseMsg(msg)

		fmt.Printf("Len: %v, err: %v\n", len(msg), err)
		if err == nil {
			t.Errorf("Len: %v, err: %v\n", len(msg), err)
		}
	}

}
