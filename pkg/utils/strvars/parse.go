package strvars

import (
	"encoding/json"
	"strings"
)

//Parse ...
func Parse(data map[string]string, v interface{}) error {
	p := make(map[string]interface{})
	for k, v := range data {
		SetVar(k, v, p)
	}
	d, _ := json.Marshal(p)
	return json.Unmarshal(d, v)
}

//SetVar ...
func SetVar(s string, value interface{}, p map[string]interface{}) {
	rs := strings.Split(s, ".")

	for i := 0; i < len(rs); i++ {
		if i == len(rs)-1 {
			p[rs[i]] = value
			return
		}
		if _, ok := p[rs[i]]; !ok {
			p[rs[i]] = make(map[string]interface{})
		}
		if _, ok := p[rs[i]].(map[string]interface{}); !ok {
			if v, ok := p[rs[i]].(map[interface{}]interface{}); ok {
				p[rs[i]] = mapconv(v)
			} else {
				p[rs[i]] = make(map[string]interface{})
			}
		}
		p = p[rs[i]].(map[string]interface{})
	}

}

func mapconv(s map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range s {
		res[k.(string)] = v
	}
	return res
}
