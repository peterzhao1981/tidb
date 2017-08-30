package session

import (
	"github.com/pingcap/tipb/go-mysqlx"
	"github.com/pingcap/tipb/go-mysqlx/Session"
	"github.com/pingcap/tidb/x-protocol/x-packetio"
)

type sessionState int32

const (
	// start as Authenticating
	authenticating         sessionState = iota
	// once authenticated, we can handle work
	ready
	// connection is closing, but wait for data to flush out first
	closing
)

type XSession struct{
	authHandler          AuthenticationHandler
	m_state              sessionState
	m_state_before_close sessionState
	session_id           uint16
	pkt                  *x_packetio.XPacketIO
}

func(xs *XSession) handleAuthMessage(msgType Mysqlx.ClientMessages_Type, payload []byte) bool {
	var r *Response
	if msgType == Mysqlx.ClientMessages_SESS_AUTHENTICATE_START {
		var data Mysqlx_Session.AuthenticateStart
		if err := data.Unmarshal(payload); err != nil {
			return false
		}

		xs.authHandler = createAuthHandler(*data.MechName, xs.pkt)
		if xs.authHandler == nil {
			xs.stop_auth()
			return false
		}

		r = xs.authHandler.handleStart(data.MechName, data.AuthData, data.InitialResponse)
	} else if msgType == Mysqlx.ClientMessages_SESS_AUTHENTICATE_CONTINUE {
		var data Mysqlx_Session.AuthenticateContinue
		if err := data.Unmarshal(payload); err != nil {
			return false
		}

		r = xs.authHandler.handleContinue(data.AuthData)
	} else {
		xs.stop_auth()
		return false
	}

	switch r.status {
	case Succeeded:
		xs.on_auth_success(r)
	case Failed:
		xs.on_auth_failure(r)
	default:
	}

	return true
}

func(xs *XSession) on_auth_success(r *Response) {
	xs.stop_auth()
	xs.m_state = ready

}

func(xs *XSession) on_auth_failure(r *Response) {
	xs.stop_auth()
}

func(xs *XSession) stop_auth() {
	xs.authHandler = nil
}
