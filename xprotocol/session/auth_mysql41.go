package session

import (
	"github.com/pingcap/tidb/mysql"
	"github.com/pingcap/tidb/util"
	"github.com/pingcap/tidb/xprotocol/x-packetio"
)

type authMysql41State int32

const (
	S_starting         authMysql41State = iota
	S_waiting_response
	S_done
	S_error
)

type saslMysql41Auth struct {
	m_state authMysql41State
	m_salt  []byte
	pkt		*x_packetio.XPacketIO
}

func (spa *saslMysql41Auth) handleStart(mechanism *string, data []byte, initial_response []byte) *Response {
	r := Response{}

	if spa.m_state == S_starting {
		spa.m_salt = util.RandomBuf(mysql.ScrambleLength)
		r.data = string(spa.m_salt)
		r.status = Ongoing
		r.errCode = 0
		spa.m_state = S_waiting_response
	} else {
		r.status = Error
		r.errCode = mysql.ErrNetPacketsOutOfOrder

		spa.m_state = S_error
	}

	return &r
}

func (spa *saslMysql41Auth) handleContinue(data []byte) *Response {
	r := Response{}

	if spa.m_state == S_waiting_response {
		//TODO check user and password here
		var err *mysql.SQLError
		err = nil
		if err == nil {
			r.status = Succeeded
			r.errCode = 0
		} else {
			r.status = Failed
			r.data = err.Message
			r.errCode = err.Code
		}
		spa.m_state = S_done
	} else {
		spa.m_state = S_error
		r.status = Error
		r.errCode = mysql.ErrNetPacketsOutOfOrder
	}

	return &r
}


