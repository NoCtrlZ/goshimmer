package tools

type StatMap map[string]interface{}

func (sn StatMap) Add(key string, val int) {
	_, ok := sn[key]
	if !ok {
		sn[key] = 0
	}
	vi, ok := sn[key].(int)
	if !ok {
		vi = 0
	}
	sn[key] = vi + val
}

func (sn StatMap) Inc(key string) {
	sn.Add(key, 1)
}

func (sn StatMap) Get(key string) interface{} {
	ret, _ := sn[key]
	return ret
}

func (sn StatMap) Set(key string, val interface{}) {
	sn[key] = val
}

func (sn StatMap) ensureInt64(key string, firstVal int64) int64 {
	var ret int64
	v, ok := sn[key]
	if ok {
		ret, ok = v.(int64)
		if ok {
			return ret
		}
	}
	sn[key] = firstVal
	return firstVal
}

func (sn StatMap) MaxInt64(key string, val int64) {
	vi := sn.ensureInt64(key, val)
	if val > vi {
		vi = val
	}
	sn[key] = vi
}

func (sn StatMap) MinInt64(key string, val int64) {
	vi := sn.ensureInt64(key, val)
	if val < vi {
		vi = val
	}
	sn[key] = vi
}
