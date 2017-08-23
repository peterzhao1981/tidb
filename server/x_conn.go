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
	"github.com/pingcap/tidb/util/arena"
)

// mysqlXClientConn represents a connection between server and client,
// it maintains connection specific state, handles client query.
type mysqlXClientConn struct {
	//pkt          *packetIO // a helper to read and write data in packet format.
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
		tp, payload, err := xcc.readPacket()
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

func (xcc *mysqlXClientConn) handshake() error {
	// TODO: implement it.
	return nil
}

// readPacket reads a full size request encoded in x protocol.
// The message struct is like:
// ______________________________________________________
// | 4 bytes length | 1 byte type | payload[0:length-1] |
// ------------------------------------------------------
// See: https://dev.mysql.com/doc/internals/en/x-protocol-messages-messages.html
func (xcc *mysqlXClientConn) readPacket() (byte, []byte, error) {
	return 0x00, nil, nil
}

func (xcc *mysqlXClientConn) dispatch(tp byte, payload []byte) error {
	return nil
}

func (xcc *mysqlXClientConn) writeError(e error) {
}
