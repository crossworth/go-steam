package netutil

import (
	"net/url"
)

func ToURLValues(m map[string]string) url.Values {
	r := make(url.Values)
	for k, v := range m {
		r.Add(k, v)
	}
	return r
}
