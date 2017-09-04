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
	"github.com/pingcap/tidb/xprotocol/xpacketio"
	"github.com/pingcap/tidb/util/arena"
	"github.com/pingcap/tipb/go-mysqlx"
	"github.com/pingcap/tidb/xprotocol/capability"
	"github.com/pingcap/tidb/xprotocol/session"
)

// mysqlXClientConn represents a connection between server and client,
// it maintains connection specific state, handles client query.
type mysqlXClientConn struct {
	pkt          *xpacketio.XPacketIO // a helper to read and write data in packet format.
	conn         net.Conn
	session      *session.XSession
	server       *Server         // a reference of server instance.
	capability   uint32          // client capability affects the way server handles client request.
	connectionID uint32          // atomically allocated by a global variable, unique in process scope.
	collation    uint8           // collation used by client, may be different from the collation used by database.
	user         string          // user of the client.
	dbname       string          // default database name.
	salt         []byte          // random bytes used for authentication.
	alloc        arena.Allocator // an memory allocator for reducing memory allocation.
	lastCmd      string          // latest sql query string, currently used for logging error.
	ctx          QueryCtx          // an interface to execute sql statements.
	attrs  map[string]string // attributes parsed from client handshake response, not used for now.
	killed bool
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
	xcc.server.rwlock.Lock()
	delete(xcc.server.clients, xcc.connectionID)
	connections := len(xcc.server.clients)
	xcc.server.rwlock.Unlock()
	connGauge.Set(float64(connections))
	xcc.conn.Close()
	//if xcc.ctx != nil {
	//	return xcc.ctx.Close()
	//}
	return nil
}

func (xcc *mysqlXClientConn) handshakeConnection() error {
	log.Infof("[YUSP] begin connection")
	tp, msg, err := xcc.pkt.ReadPacket()
	if err != nil {
		return errors.Trace(err)
	}
	log.Infof("[YUSP] deal first msg")
	if err = capability.DealInitCapabilitiesSet(Mysqlx.ClientMessages_Type(tp), msg); err != nil {
		return errors.Trace(err)
	}
	log.Infof("[YUSP] send first msg")
	if err = xcc.pkt.WritePacket(int32(Mysqlx.ServerMessages_OK), []byte{}); err != nil {
		return errors.Trace(err)
	}
	log.Infof("[YUSP] read sec msg")
	tp, msg, err = xcc.pkt.ReadPacket()
	if err != nil {
		return errors.Trace(err)
	}
	log.Infof("[YUSP] deal sec msg")
	if err = capability.DealCapabilitiesGet(Mysqlx.ClientMessages_Type(tp), msg); err != nil {
		return errors.Trace(err)
	}
	resp, err := capability.GetCapabilities().Marshal()
	if err != nil {
		return errors.Trace(err)
	}
	if err = xcc.pkt.WritePacket(int32(Mysqlx.ServerMessages_CONN_CAPABILITIES), resp); err != nil {
		return errors.Trace(err)
	}
	tp, msg, err = xcc.pkt.ReadPacket()
	if err != nil {
		return errors.Trace(err)
	}
	if err = capability.DealSecCapabilitiesSet(Mysqlx.ClientMessages_Type(tp), msg); err != nil {
		return errors.Trace(err)
	}
	resp, err = capability.CapabilityErrorReport().Marshal()
	if err != nil {
		return errors.Trace(err)
	}
	if err = xcc.pkt.WritePacket(int32(Mysqlx.ServerMessages_ERROR), resp); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (xcc *mysqlXClientConn) handshakeSession() error {
	xcc.session = session.CreateSession(xcc.id(), xcc.pkt)
	tp, msg, err := xcc.pkt.ReadPacket()
	if err != nil {
		return errors.Trace(err)
	}

	if err := xcc.session.HandleAuthMessage(Mysqlx.ClientMessages_Type(tp), msg); err != nil {
		return errors.New("error happened when handle auth start.")
	}

	tp, msg, err = xcc.pkt.ReadPacket()
	if err != nil {
		return errors.Trace(err)
	}

	if err := xcc.session.HandleAuthMessage(Mysqlx.ClientMessages_Type(tp), msg); err != nil {
		return errors.New("error happened when handle auth continue.")
	}

	return nil
}

func (xcc *mysqlXClientConn) handshake() error {
	if err := xcc.handshakeConnection(); err != nil {
		return err
	}

	if err := xcc.handshakeSession(); err != nil {
		return err
	}

	return nil
}

func (xcc *mysqlXClientConn) dispatch(tp int32, payload []byte) error {
	if err := xcc.session.HandleReadyMessage(Mysqlx.ClientMessages_Type(tp), payload); err != nil {
		return errors.New("dispatch error")
	}
	return nil
}

func (xcc *mysqlXClientConn) flush() error {
	return xcc.pkt.Flush()
}
func (xcc *mysqlXClientConn) writeError(e error) {
}

func (xcc *mysqlXClientConn) isKilled() bool {
	return xcc.killed
}

func (xcc *mysqlXClientConn) Cancel(query bool) {
	//xcc.ctx.Cancel()
	if !query {
		xcc.killed = true
	}
}

func (xcc *mysqlXClientConn) id() uint32 {
	return xcc.connectionID
}

func (xcc *mysqlXClientConn) showProcess() util.ProcessInfo {
	//return xcc.ctx.ShowProcess()
	return util.ProcessInfo{}
}
