package generic

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
	"sort"
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
	var ret int
	switch rett := ret1.(type) {
	case int:
		ret = rett
	case int16:
		ret = int(rett)
	case uint16:
		ret = int(rett)
	case int32:
		ret = int(rett)
	case uint32:
		ret = int(rett)
	case int64:
		ret = int(rett)
	case uint64:
		ret = int(rett)
	case float32:
		ret = int(rett)
	case float64:
		ret = int(rett)
	}
	vm[name] = ret
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

// to achieve deterministic json data of the dictionary
// it must be sorted first

func (vm flatValueMap) Write(w io.Writer) error {
	keys, values := vm.sortedMap()
	dataK, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	dataV, err := json.Marshal(values)
	if err != nil {
		return err
	}
	err = tools.WriteBytes32(w, dataK)
	if err != nil {
		return err
	}
	err = tools.WriteBytes32(w, dataV)
	if err != nil {
		return err
	}
	return err
}

func (vm flatValueMap) Read(r io.Reader) error {
	dataK, err := tools.ReadBytes32(r)
	if err != nil {
		return err
	}
	keys := make([]string, 0)
	err = json.Unmarshal(dataK, &keys)
	if err != nil {
		return err
	}
	dataV, err := tools.ReadBytes32(r)
	if err != nil {
		return err
	}
	values := make([]interface{}, 0)
	err = json.Unmarshal(dataV, &values)
	if err != nil {
		return err
	}
	return vm.reconstruct(keys, values)
}

func (vm flatValueMap) sortedMap() ([]string, []interface{}) {
	retK := make([]string, 0, len(vm))
	for k := range vm {
		retK = append(retK, string(k))
	}
	sort.Strings(retK)
	retV := make([]interface{}, 0, len(vm))
	for _, k := range retK {
		retV = append(retV, vm[VarName(k)])
	}
	return retK, retV
}

func (vm flatValueMap) reconstruct(keys []string, values []interface{}) error {
	if len(keys) != len(values) {
		return fmt.Errorf("len(keys) != len(values)")
	}
	for i := range keys {
		vm[VarName(keys[i])] = values[i]
	}
	return nil
}

// shallow copy of values

func (vm flatValueMap) Clone() ValueMap {
	ret := make(flatValueMap)
	for k, v := range vm {
		ret[k] = v
	}
	return ret
}
