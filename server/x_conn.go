// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"io"
	"net"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pingcap/tidb/terror"
	"github.com/pingcap/tidb/util"
	"github.com/pingcap/tidb/x-protocol/x-packetio"
	"github.com/pingcap/tidb/util/arena"
	"github.com/pingcap/tipb/go-mysqlx"
	"github.com/pingcap/tipb/go-mysqlx/Connection"
	"github.com/pingcap/tipb/go-mysqlx/Session"
	"github.com/pingcap/tipb/go-mysqlx/Sql"
)

// mysqlXClientConn represents a connection between server and client,
// it maintains connection specific state, handles client query.
type mysqlXClientConn struct {
	pkt          *x_packetio.XPacketIO // a helper to read and write data in packet format.
	conn         net.Conn
	server       *Server           // a reference of server instance.
	capability   uint32            // client capability affects the way server handles client request.
	connectionID uint32            // atomically allocated by a global variable, unique in process scope.
	collation    uint8             // collation used by client, may be different from the collation used by database.
	user         string            // user of the client.
	dbname       string            // default database name.
	salt         []byte            // random bytes used for authentication.
	alloc        arena.Allocator   // an memory allocator for reducing memory allocation.
	lastCmd      string            // latest sql query string, currently used for logging error.
	//ctx          QueryCtx          // an interface to execute sql statements.
	attrs        map[string]string // attributes parsed from client handshake response, not used for now.
	killed       bool
}

func (xcc *mysqlXClientConn) Run() {
	defer func() {
		recover()
		xcc.Close()
	}()

	for !xcc.killed {
		tp, payload, err := xcc.pkt.ReadPacket()
		if err != nil {
			if terror.ErrorNotEqual(err, io.EOF) {
				log.Errorf("[%d] read packet error, close this connection %s",
					xcc.connectionID, errors.ErrorStack(err))
			}
			return
		}
		if err = xcc.dispatch(tp, payload); err != nil {
			if terror.ErrorEqual(err, terror.ErrResultUndetermined) {
				log.Errorf("[%d] result undetermined error, close this connection %s",
					xcc.connectionID, errors.ErrorStack(err))
			} else if terror.ErrorEqual(err, terror.ErrCritical) {
				log.Errorf("[%d] critical error, stop the server listener %s",
					xcc.connectionID, errors.ErrorStack(err))
				select {
				case xcc.server.stopListenerCh <- struct{}{}:
				default:
				}
			}
			log.Warnf("[%d] dispatch error: %s, %s", xcc.connectionID, xcc, err)
			xcc.writeError(err)
			return
		}
	}
}

func (xcc *mysqlXClientConn) Close() error {
	err := xcc.conn.Close()
	return errors.Trace(err)
}


func (xcc *mysqlXClientConn) handshakeConnection() error {

	return nil
}

func (xcc *mysqlXClientConn) handshakeSession() error {
	return nil
}

func (xcc *mysqlXClientConn) handshake() error {
	if err := xcc.handshakeConnection(); err != nil {

	}

	if err := xcc.handshakeSession(); err != nil {

	}

	return nil
}

func (xcc *mysqlXClientConn) dispatch(tp int32, payload []byte) error {
	switch Mysqlx.ClientMessages_Type(tp) {
	case Mysqlx.ClientMessages_CON_CAPABILITIES_GET:
		var data Mysqlx_Connection.CapabilitiesGet
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_CON_CAPABILITIES_SET:
		var data Mysqlx_Connection.CapabilitiesSet
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_CON_CLOSE:
		var data Mysqlx_Connection.Close
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_SESS_AUTHENTICATE_START:
		var data Mysqlx_Session.AuthenticateStart
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_SESS_AUTHENTICATE_CONTINUE:
		var data Mysqlx_Session.AuthenticateContinue
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_SESS_RESET:
		var data Mysqlx_Session.Reset
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_SESS_CLOSE:
		var data Mysqlx_Session.Close
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_SQL_STMT_EXECUTE:
		var data Mysqlx_Sql.StmtExecute
		if err := data.Unmarshal(payload); err != nil {
			return err
		}
	case Mysqlx.ClientMessages_CRUD_FIND:
	case Mysqlx.ClientMessages_CRUD_INSERT:
	case Mysqlx.ClientMessages_CRUD_UPDATE:
	case Mysqlx.ClientMessages_CRUD_DELETE:
	case Mysqlx.ClientMessages_EXPECT_OPEN:
	case Mysqlx.ClientMessages_EXPECT_CLOSE:
	case Mysqlx.ClientMessages_CRUD_CREATE_VIEW:
	case Mysqlx.ClientMessages_CRUD_MODIFY_VIEW:
	case Mysqlx.ClientMessages_CRUD_DROP_VIEW:
	}
	return nil
}

func (xcc *mysqlXClientConn) writeError(e error) {
}


func (cc *mysqlXClientConn) isKilled() bool {
	return cc.killed
}

func (cc *mysqlXClientConn) Cancel(query bool) {
	//cc.ctx.Cancel()
	if !query {
		cc.killed = true
	}
}

func (cc *mysqlXClientConn) id() uint32 {
	return cc.connectionID
}

func (cc *mysqlXClientConn) showProcess() util.ProcessInfo {
	//return cc.ctx.ShowProcess()
	return util.ProcessInfo{}
}
