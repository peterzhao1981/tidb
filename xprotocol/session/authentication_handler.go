package session

import (
	"github.com/ngaut/log"
	"github.com/pingcap/tidb/xprotocol/x-packetio"
)

type Status int32

const (
	Ongoing   Status = iota
	Succeeded
	Failed
	Error
)

type Response struct {
	data    string
	status  Status
	errCode uint16
}

type AuthenticationHandler interface {
	handleStart(mechanism *string, data []byte, initial_response []byte) *Response
	handleContinue(data []byte) *Response
}

func createAuthHandler(method string, pkt *x_packetio.XPacketIO) AuthenticationHandler {
	switch method {
	case "MYSQL41":
		return &saslMysql41Auth{
			m_state:S_starting,
			pkt: pkt,
		}
	case "PLAIN":
		return &saslPlainAuth{}
	default:
		log.Error("unknown auth handler type.")
		return nil
	}
}
