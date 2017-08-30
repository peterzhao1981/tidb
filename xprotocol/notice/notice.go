package notice

import (
	"github.com/pingcap/tipb/go-mysqlx/Notice"
)

type Notice struct {
}

const (
	k_notice_warning                  uint32 = 1
	k_notice_session_variable_changed        = 2
	k_notice_session_state_changed           = 3
)

func SendLocalNotice(noticeType uint32, data *string, forceFlush bool) error {
	return nil
	//return p.Send_notice(noticeType, data, Mysqlx_Notice.Frame_LOCAL, forceFlush)
}

func SendNotice(noticeTyp uint32, data *string, scope Mysqlx_Notice.Frame_Scope, forceFlush bool) error {
	noticeMsg := new(Mysqlx_Notice.Frame)

	switch scope {
	case Mysqlx_Notice.SessionStateChanged_CURRENT_SCHEMA, Mysqlx_Notice.SessionStateChanged_ACCOUNT_EXPIRED,
		Mysqlx_Notice.SessionStateChanged_GENERATED_INSERT_ID, Mysqlx_Notice.SessionStateChanged_ROWS_AFFECTED,
		Mysqlx_Notice.SessionStateChanged_ROWS_FOUND, Mysqlx_Notice.SessionStateChanged_ROWS_MATCHED,
		Mysqlx_Notice.SessionStateChanged_TRX_COMMITTED, Mysqlx_Notice.SessionStateChanged_TRX_ROLLEDBACK,
		Mysqlx_Notice.SessionStateChanged_PRODUCED_MESSAGE, Mysqlx_Notice.SessionStateChanged_CLIENT_ID_ASSIGNED:

		noticeMsg.Type = &uint32(k_notice_session_state_changed)
		//msg := Mysqlx_Notice.SessionStateChanged{
		//	Param: &Mysqlx_Notice.SessionStateChanged_Parameter(noticeTyp),
		//	Value:
		//}
	default:
	}

	//notice := Mysqlx_Notice.Frame{}
	return nil
}
