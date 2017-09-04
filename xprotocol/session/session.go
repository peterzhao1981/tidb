package session

import (
	"github.com/juju/errors"
	"github.com/pingcap/tipb/go-mysqlx"
	"github.com/pingcap/tipb/go-mysqlx/Session"
	"github.com/pingcap/tidb/xprotocol/xpacketio"
	"github.com/pingcap/tidb/xprotocol/notice"
	"github.com/pingcap/tidb/xprotocol/util"
	"github.com/pingcap/tidb/mysql"
	"github.com/pingcap/tipb/go-mysqlx/Sql"
)

type sessionState int32

const (
	// start as Authenticating
	authenticating sessionState = iota
	// once authenticated, we can handle work
	ready
	// connection is closing, but wait for data to flush out first
	closing
)

type XSession struct {
	authHandler       AuthenticationHandler
	mState            sessionState
	mStateBeforeClose sessionState
	sessionId         uint32
	pkt               *xpacketio.XPacketIO
}

func (xs *XSession) handleMessage(msgType Mysqlx.ClientMessages_Type, payload []byte) error {
	if xs.mState == authenticating {
		return xs.HandleAuthMessage(msgType, payload)
	} else if xs.mState == ready {
		return xs.HandleReadyMessage(msgType, payload)
	}

	return errors.New("unknown session state")
}

func (xs *XSession) HandleReadyMessage(msgType Mysqlx.ClientMessages_Type, payload []byte) error {
	switch msgType {
	case Mysqlx.ClientMessages_SESS_CLOSE:
		content := "bye!"
		notice.SendOK(xs.pkt, &content)
		xs.onClose(false)
		return nil
	case Mysqlx.ClientMessages_CON_CLOSE:
		content := "bye!"
		notice.SendOK(xs.pkt, &content)
		xs.onClose(false)
		return nil
	case Mysqlx.ClientMessages_SESS_RESET:
		xs.mState = closing
		xs.onSessionReset()
		return nil
	case Mysqlx.ClientMessages_SQL_STMT_EXECUTE:
		var data Mysqlx_Sql.StmtExecute
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
		if err := xs.DealSQLStmtExecute(data); err != nil {
			return err
		}

	}
	return errors.New("invalid message type")
}

func (xs *XSession) HandleAuthMessage(msgType Mysqlx.ClientMessages_Type, payload []byte) error {
	var r *Response
	if msgType == Mysqlx.ClientMessages_SESS_AUTHENTICATE_START {
		var data Mysqlx_Session.AuthenticateStart
		if err := data.Unmarshal(payload); err != nil {
			errCode := util.ErXBadMessage
			content := "Invalid message"
			notice.SendInitError(xs.pkt, &errCode, &content)
			return err
		}

		xs.authHandler = createAuthHandler(*data.MechName, xs.pkt)
		if xs.authHandler == nil {
			errCode := uint16(mysql.ErrNotSupportedAuthMode)
			content := "Invalid authentication method " + *data.MechName
			notice.SendInitError(xs.pkt, &errCode, &content)
			xs.stopAuth()
			return errors.New("invalid authentication method")
		}

		r = xs.authHandler.handleStart(data.MechName, data.AuthData, data.InitialResponse)
	} else if msgType == Mysqlx.ClientMessages_SESS_AUTHENTICATE_CONTINUE {
		var data Mysqlx_Session.AuthenticateContinue
		if err := data.Unmarshal(payload); err != nil {
			errCode := util.ErXBadMessage
			content := "Invalid message"
			notice.SendInitError(xs.pkt, &errCode, &content)
			return err
		}

		r = xs.authHandler.handleContinue(data.AuthData)
	} else {
		errCode := util.ErXBadMessage
		content := "Invalid message"
		notice.SendInitError(xs.pkt, &errCode, &content)
		xs.stopAuth()
		return errors.New("invalid message")
	}

	switch r.status {
	case Succeeded:
		xs.onAuthSuccess(r)
	case Failed:
		xs.onAuthFailure(r)
	default:
		xs.SendAuthContinue(&r.data)
	}

	return nil
}

func (xs *XSession) onAuthSuccess(r *Response) {
	notice.SendClientId(xs.pkt, xs.sessionId)
	xs.stopAuth()
	xs.mState = ready
	xs.SendAuthOk(&r.data)

}

func (xs *XSession) onAuthFailure(r *Response) {
	errCode := uint16(mysql.ErrAccessDenied)
	notice.SendInitError(xs.pkt, &errCode, &r.data)
	xs.stopAuth()
}

//@TODO need to implement
func (xs *XSession) onSessionReset() {
}

func (xs *XSession) onClose(updateOldState bool) {
	if xs.mState != closing {
		if updateOldState {
			xs.mStateBeforeClose = xs.mState
		}
		xs.mState = closing
	}
}

func (xs *XSession) stopAuth() {
	xs.authHandler = nil
}

func (xs *XSession) SendAuthOk(value *string) error {
	msg := Mysqlx_Session.AuthenticateOk{
		AuthData: []byte(*value),
	}

	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	return xs.pkt.WritePacket(int32(Mysqlx.ServerMessages_SESS_AUTHENTICATE_OK), data)
}

func (xs *XSession) SendAuthContinue(value *string) error {
	msg := Mysqlx_Session.AuthenticateContinue{
		AuthData: []byte(*value),
	}

	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	return xs.pkt.WritePacket(int32(Mysqlx.ServerMessages_SESS_AUTHENTICATE_CONTINUE), data)
}

func CreateSession(id uint32, pkt *xpacketio.XPacketIO) *XSession {
	return &XSession{
		mState: authenticating,
		mStateBeforeClose: authenticating,
		sessionId: id,
		pkt: pkt,
	}
}

func (xs *XSession) DealSQLStmtExecute (msg Mysqlx_Sql.StmtExecute) error {
	switch msg.GetNamespace() {
	case "xplugin":
	case "mysqlx":
	case "sql", "":
		//sql := string(msg.GetStmt())
	default:
		return errors.New("unknown namespace")
	}
	return nil
}
