package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/thorory/lorawan_decode/internel/decode"
	"github.com/thorory/lorawan_decode/internel/ui"
)

var phyPayload string

func init() {
	flag.StringVar(&phyPayload, "d", "", "PHYPayload of LoRaWAN packet")
}

func main() {
	flag.Parse()

	ui.InitClipboard()

	b, err := hex.DecodeString(phyPayload)
	if err != nil {
		fmt.Printf("decode error %s\n", err.Error())
		os.Exit(-1)
	}
	if content, err := decode.PHYPayloadMarshalToText(b); err != nil {
		fmt.Printf("decode error %s\n", err.Error())
		os.Exit(-1)
	} else {
		fmt.Println(content)
	}

	os.Exit(1)
}
