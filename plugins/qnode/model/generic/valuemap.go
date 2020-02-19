package generic

import (
	"encoding/json"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type VarName string

type ValueMap interface {
	GetInt(VarName) (int, bool)
	GetString(VarName) (string, bool)
	GetFloat64(VarName) (float64, bool)
	GetBool(name VarName) (bool, bool)
	SetInt(VarName, int)
	SetString(VarName, string)
	SetFloat64(VarName, float64)
	SetBool(name VarName, value bool)
	Clone() ValueMap
	Encode() Encode
}

type flatValueMap map[VarName]interface{}

func NewFlatValueMap() ValueMap {
	return make(flatValueMap)
}

func (vm flatValueMap) GetBool(name VarName) (bool, bool) {
	ret1, ok := vm[name]
	if !ok {
		return false, false
	}
	ret, ok := ret1.(bool)
	if !ok {
		return false, false
	}
	return ret, true
}

func (vm flatValueMap) GetInt(name VarName) (int, bool) {
	ret1, ok := vm[name]
	if !ok {
		return 0, false
	}
	ret, ok := ret1.(int)
	if !ok {
		return 0, false
	}
	return ret, true
}

func (vm flatValueMap) GetString(name VarName) (string, bool) {
	ret1, ok := vm[name]
	if !ok {
		return "", false
	}
	ret, ok := ret1.(string)
	if !ok {
		return "", false
	}
	return ret, true
}

func (vm flatValueMap) GetFloat64(name VarName) (float64, bool) {
	ret1, ok := vm[name]
	if !ok {
		return 0, false
	}
	ret, ok := ret1.(float64)
	if !ok {
		return 0, false
	}
	return ret, true
}

func (vm flatValueMap) SetBool(name VarName, value bool) {
	vm[name] = value
}

func (vm flatValueMap) SetInt(name VarName, val int) {
	vm[name] = val
}

func (vm flatValueMap) SetString(name VarName, val string) {
	vm[name] = val
}

func (vm flatValueMap) SetFloat64(name VarName, val float64) {
	vm[name] = val
}

func (vm flatValueMap) Encode() Encode {
	return vm
}

func (vm flatValueMap) Write(w io.Writer) error {
	data, err := json.Marshal(vm)
	if err != nil {
		return err
	}
	err = tools.WriteBytes32(w, data)
	return err
}

func (vm flatValueMap) Read(r io.Reader) error {
	data, err := tools.ReadBytes32(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &vm)
}

// shallow copy of values

func (vm flatValueMap) Clone() ValueMap {
	ret := make(flatValueMap)
	for k, v := range vm {
		ret[k] = v
	}
	return ret
}
