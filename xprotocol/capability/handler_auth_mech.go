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

package capability

import (
	"github.com/pingcap/tipb/go-mysqlx/Connection"
	"github.com/pingcap/tipb/go-mysqlx/Datatypes"
	"github.com/pingcap/tidb/xprotocol/util"
)

type HandleAuthMech struct {
	Values []string
}

func (h *HandleAuthMech) IsSupport() bool {
	return true
}

func (h *HandleAuthMech) GetName() string {
	return "authentication.mechanisms"
}

func (h *HandleAuthMech) Get() *Mysqlx_Connection.Capability {
	meths := []Mysqlx_Datatypes.Any{}
	for _, v := range h.Values {
		meths = append(meths, util.SetString([]byte(v)))
	}
	val := util.SetScalarArray(meths)
	str := new(string)
	*str = h.GetName()
	c := Mysqlx_Connection.Capability{
		Name: str,
		Value: &val,
	}
	return &c
}

func (h *HandleAuthMech) Set(any *Mysqlx_Datatypes.Any) bool {
	return false
}
