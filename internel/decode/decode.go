package decode

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/brocaar/lorawan"
)

type Content struct {
	PHYPayload PHYPayload
	RawPacket  []byte
}

// PHYPayload contains the decoded phypayload of lorawan packet
type PHYPayload struct {
	MHDR        MHDR         `json:"MHDR"`
	MACPayload  *MACPayload  `json:"MACPayload,omitempty"`
	JoinRequest *JoinRequest `json:"JoinRequest,omitempty"`
	JoinAccept  *JoinAccept  `json:"JoinAccept,omitempty"`
	MIC         string       `json:"MIC"`
}

// MHDR contains the decoded MHDR of lorawan packet
type MHDR struct {
	MType string `json:"MType"`
	RFU   string `json:"RFU"`
	Major string `json:"Major"`
}

// MACPayload contains the decoded MACPayload of lorawan packet
type MACPayload struct {
	FHDR       FHDR   `json:"FHDR"`
	FPort      uint32 `json:"FPort"`
	FRMPayload string `json:"FRMPayload"`
}

// FHDR contains the decoded FHDR of lorawan packet
type FHDR struct {
	DevAddr string `json:"DevEUI"`
	FCtrl   FCtrl  `json:"FCtrl"`
	FCnt    uint32 `json:"FCnt"`
	FOpts   FOpts  `json:"FOpts"`
}

// FCtrl contains the decoded FCtrl of lorawan packet
type FCtrl struct {
	ADR       bool    `json:"ADR"`
	RFU       *string `json:"RFU,omitempty"`
	ADRACKReq *bool   `json:"ADRACKReq,omitempty"`
	ACK       bool    `json:"ACK"`
	FPending  *bool   `json:"FPending,omitempty"`
	ClassB    *bool   `json:"ClassB,omitempty"`
	FOptsLen  uint32  `json:"FoptsLen"`
}

type FOpts struct {
	Payload string `json:"Payload,omitempty"`
	// Maccommands string
}

// JoinRequest contains the decoded JoinRequest of lorawan packet
type JoinRequest struct {
	AppEUI   string `json:"AppEUI"`
	DevEUI   string `json:"DevEUI"`
	DevNonce string `json:"DevNonce"`
}

// JoinAccept contains the decoded JoinAccept of lorawan packet
type JoinAccept struct {
	AppNonce  string  `json:"AppNonce"`
	NetID     string  `json:"NetID"`
	DevAddr   string  `json:"DevAddr"`
	DLSetting string  `json:"DLSetting"`
	RxDelay   uint8   `json:"RxDelay"`
	CFList    *string `json:"CFList,omitempty"`
}

// PHYPayloadMarshalToText decode the lorawan PHYPayload into text
func PHYPayloadMarshalToText(phypayload []byte) (string, error) {
	content := &Content{
		RawPacket: phypayload,
	}
	p := &lorawan.PHYPayload{}
	err := p.UnmarshalBinary(phypayload)
	if err != nil {
		return "", err
	}

	// decode phypayload
	err = content.decodePHYPayload(p)
	if err != nil {
		return "", err
	}

	return content.String(), nil
}

func (c *Content) String() string {

	raw := hex.EncodeToString(c.RawPacket)
	pkt, _ := json.MarshalIndent(c.PHYPayload, "", "    ")

	return fmt.Sprintf("Raw PHYPayload: %s\nContent: %s\n", raw, pkt)
}

func (c *Content) decodePHYPayload(p *lorawan.PHYPayload) error {
	c.PHYPayload.MHDR.MType = p.MHDR.MType.String()
	c.PHYPayload.MHDR.RFU = fmt.Sprintf("%b", ((c.RawPacket[0] & 0x1c) >> 2))
	c.PHYPayload.MHDR.Major = p.MHDR.Major.String()
	c.PHYPayload.MIC = p.MIC.String()

	var err error

	switch p.MHDR.MType {
	case lorawan.JoinRequest:
		joinRequest, err := decodeJoinRequest(p)
		c.PHYPayload.JoinRequest = &joinRequest
		return err
	case lorawan.JoinAccept:
		joinAccept, err := decodeJoinAccept(p)
		c.PHYPayload.JoinAccept = &joinAccept
		return err
	case lorawan.UnconfirmedDataUp, lorawan.ConfirmedDataUp:
		macpayload, err := decodeUplinkMACPayload(p)
		c.PHYPayload.MACPayload = &macpayload
		return err
	case lorawan.UnconfirmedDataDown, lorawan.ConfirmedDataDown:
		macpayload, err := decodeDownlinkMACPayload(p)
		c.PHYPayload.MACPayload = &macpayload
		return err
	case lorawan.RejoinRequest:
		err = errors.New("not support rejoin_request packet")
		return err
	case lorawan.Proprietary:
		err = errors.New("not support proprietary packet")
		return err
	default:
		err = errors.New("unknown packet")
		return err
	}
}

func decodeJoinRequest(p *lorawan.PHYPayload) (JoinRequest, error) {
	joinRequest, ok := p.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		return JoinRequest{}, errors.New("lorawan: MACPayload value must be of type *JoinRequestPayload")
	}

	devNonce, _ := joinRequest.DevNonce.MarshalBinary()
	return JoinRequest{
		AppEUI:   joinRequest.JoinEUI.String(),
		DevEUI:   joinRequest.DevEUI.String(),
		DevNonce: hex.EncodeToString(devNonce),
	}, nil
}

func decodeJoinAccept(p *lorawan.PHYPayload) (JoinAccept, error) {
	joinAccept, ok := p.MACPayload.(*lorawan.JoinAcceptPayload)
	if !ok {
		return JoinAccept{}, errors.New("lorawan: MACPayload value must be of type *JoinAcceptPayload")
	}

	appNonce, _ := joinAccept.JoinNonce.MarshalBinary()
	DLSetting, _ := joinAccept.DLSettings.MarshalBinary()
	var cfList *string
	if joinAccept.CFList != nil {
		bList, _ := joinAccept.CFList.MarshalBinary()
		sList := hex.EncodeToString(bList)
		cfList = &sList
	}
	return JoinAccept{
		AppNonce:  hex.EncodeToString(appNonce),
		NetID:     joinAccept.HomeNetID.String(),
		DevAddr:   joinAccept.DevAddr.String(),
		DLSetting: hex.EncodeToString(DLSetting),
		RxDelay:   joinAccept.RXDelay,
		CFList:    cfList,
	}, nil
}

func decodeUplinkMACPayload(p *lorawan.PHYPayload) (MACPayload, error) {
	m, ok := p.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return MACPayload{}, errors.New("lorawan: MACPayload value must be of type *MACPayload")
	}

	var fport uint32
	if m.FPort != nil {
		fport = uint32(*m.FPort)
	}

	var fopts []byte
	for _, fopt := range m.FHDR.FOpts {
		b, _ := fopt.MarshalBinary()
		fopts = append(fopts, b...)
	}

	var frmPayload []byte
	for _, payload := range m.FRMPayload {
		b, _ := payload.MarshalBinary()
		frmPayload = append(frmPayload, b...)
	}

	return MACPayload{
		FPort: fport,
		FHDR: FHDR{
			DevAddr: m.FHDR.DevAddr.String(),
			FCnt:    m.FHDR.FCnt,
			FCtrl: FCtrl{
				ADR:       m.FHDR.FCtrl.ADR,
				ADRACKReq: &m.FHDR.FCtrl.ADRACKReq,
				ACK:       m.FHDR.FCtrl.ACK,
				ClassB:    &m.FHDR.FCtrl.ClassB,
			},
			FOpts: FOpts{
				Payload: hex.EncodeToString(fopts), // TODO
			},
		},
		FRMPayload: hex.EncodeToString(frmPayload),
	}, nil
}

func decodeDownlinkMACPayload(p *lorawan.PHYPayload) (MACPayload, error) {
	m, ok := p.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return MACPayload{}, errors.New("lorawan: MACPayload value must be of type *MACPayload")
	}

	var fport uint32
	if m.FPort != nil {
		fport = uint32(*m.FPort)
	}

	var fopts []byte
	for _, fopt := range m.FHDR.FOpts {
		b, _ := fopt.MarshalBinary()
		fopts = append(fopts, b...)
	}

	var frmPayload []byte
	for _, payload := range m.FRMPayload {
		b, _ := payload.MarshalBinary()
		frmPayload = append(frmPayload, b...)
	}

	return MACPayload{
		FPort: fport,
		FHDR: FHDR{
			DevAddr: m.FHDR.DevAddr.String(),
			FCnt:    m.FHDR.FCnt,
			FCtrl: FCtrl{
				ADR:      m.FHDR.FCtrl.ADR,
				ACK:      m.FHDR.FCtrl.ACK,
				FPending: &m.FHDR.FCtrl.FPending,
			},
			FOpts: FOpts{
				Payload: hex.EncodeToString(fopts), // TODO
			},
		},
		FRMPayload: hex.EncodeToString(frmPayload),
	}, nil
}
