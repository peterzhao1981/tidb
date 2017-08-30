package xprotocol

import (
	"github.com/pingcap/tipb/go-mysqlx/Connection"
	"github.com/pingcap/tidb/xprotocol/capability"
)

func GetCapabilites() Mysqlx_Connection.Capabilities {
	authHandler := capability.HandleAuthMech {
		Values: []string{"MYSQL41"},
	}
	docHandler := capability.HandlerReadOnlyValue{
		Name: "doc.formats",
		Value: "text",
	}
	nodeHandler := capability.HandlerReadOnlyValue{
		Name: "node_type",
		Value: "mysql",
	}
	pwdHandler := capability.HandlerExpiredPasswords{
		Name: "client.pwd_expire_ok",
		Expired: true,
	}
	caps := Mysqlx_Connection.Capabilities{
		Capabilities: []*Mysqlx_Connection.Capability{
			&authHandler.Get(),
			&docHandler.Get(),
			&nodeHandler.Get(),
			&pwdHandler.Get(),
		},
	}
	return caps
}
