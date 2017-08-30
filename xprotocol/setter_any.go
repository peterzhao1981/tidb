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

package mysqlx

import (
	"github.com/pingcap/tipb/go-mysqlx/Datatypes"
)

func scalar2Any(scalar Mysqlx_Datatypes.Scalar) Mysqlx_Datatypes.Any {
	any := Mysqlx_Datatypes.Any{
		Type: Mysqlx_Datatypes.Any_SCALAR.Enum(),
		Scalar: &scalar,
	}
	return any
}

func SetSint(data int64) Mysqlx_Datatypes.Any {
	sca := Mysqlx_Datatypes.Scalar {
		Type: Mysqlx_Datatypes.Scalar_V_SINT.Enum(),
		VSignedInt: &data,
	}
	return scalar2Any(sca)
}

func SetUint(data uint64) Mysqlx_Datatypes.Any {
	sca := Mysqlx_Datatypes.Scalar {
		Type: Mysqlx_Datatypes.Scalar_V_UINT.Enum(),
		VUnsignedInt: &data,
	}
	return scalar2Any(sca)
}

func SetOct(data []byte) Mysqlx_Datatypes.Any {
	oct := Mysqlx_Datatypes.Scalar_Octets{
		Value: data,
	}
	sca := Mysqlx_Datatypes.Scalar {
		Type: Mysqlx_Datatypes.Scalar_V_OCTETS.Enum(),
		VOctets: &oct,
	}
	return scalar2Any(sca)
}

func SetDouble(data float64) Mysqlx_Datatypes.Any {
	sca := Mysqlx_Datatypes.Scalar {
		Type: Mysqlx_Datatypes.Scalar_V_DOUBLE.Enum(),
		VDouble: &data,
	}
	return scalar2Any(sca)
}

func SetFloat(data float32) Mysqlx_Datatypes.Any {
	sca := Mysqlx_Datatypes.Scalar {
		Type: Mysqlx_Datatypes.Scalar_V_FLOAT.Enum(),
		VFloat: &data,
	}
	return scalar2Any(sca)
}

func SetBool(data bool) Mysqlx_Datatypes.Any {
	sca := Mysqlx_Datatypes.Scalar {
		Type: Mysqlx_Datatypes.Scalar_V_BOOL.Enum(),
		VBool: &data,
	}
	return scalar2Any(sca)
}

func SetString(data []byte) Mysqlx_Datatypes.Any {
	str := Mysqlx_Datatypes.Scalar_String{
		Value: data,
	}
	sca := Mysqlx_Datatypes.Scalar {
		Type: Mysqlx_Datatypes.Scalar_V_STRING.Enum(),
		VString: &str,
	}
	return scalar2Any(sca)
}

func SetScalarArray(anys []Mysqlx_Datatypes.Any) Mysqlx_Datatypes.Any {
	vals := []*Mysqlx_Datatypes.Any{}
	for _, v := range anys {
		vals = append(vals, &v)
	}
	arr := Mysqlx_Datatypes.Array {
		Value: vals,
	}
	val := Mysqlx_Datatypes.Any {
		Type: Mysqlx_Datatypes.Any_ARRAY.Enum(),
		Array: &arr,
	}
	return val
}
