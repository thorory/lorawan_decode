package ui

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/thorory/lorawan_decode/internel/decode"
)

func InitClipboard() {
	var inTE, outTE *walk.TextEdit

	MainWindow{
		Title:   "SCREAMO",
		MinSize: Size{1200, 800},
		Layout:  VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					TextEdit{AssignTo: &inTE},
					TextEdit{
						AssignTo: &outTE,
						HScroll:  true,
						VScroll:  true,
						ReadOnly: true,
					},
				},
			},
			PushButton{
				Text: "SCREAM",
				OnClicked: func() {
					run(inTE, outTE)
				},
			},
		},
	}.Run()
}

func run(inTe *walk.TextEdit, outTe *walk.TextEdit) {
	b, err := hex.DecodeString(inTe.Text())
	if err != nil {
		outTe.SetText(fmt.Sprintf("decode error %s\n", err.Error()))
	}
	if content, err := decode.PHYPayloadMarshalToText(b); err != nil {
		outTe.SetText(fmt.Sprintf("decode error %s\n", err.Error()))
	} else {
		content = strings.Replace(content, "\r\n", "\n", -1)
		content = strings.Replace(content, "\n", "\r\n", -1)
		outTe.SetText(content)
	}
}
