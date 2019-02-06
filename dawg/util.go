package dawg

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// URLParam is an interface that represents a url parameter. It was defined
// so that url.Params from "net/url" can also be used
type URLParam interface {
	Encode() string
}

// Params represents parameters passed to a url
type Params map[string]interface{}

// Encode converts the map alias to a string representation of a url parameter.
func (p Params) Encode() string {
	// I totally stole this function from the net/url parckage. I should probably
	// give credit where it is due.
	if p == nil {
		return ""
	}
	var val string
	var buffer strings.Builder
	for k, v := range p {
		key := url.QueryEscape(k)
		switch v.(type) {
		case int:
			val = strconv.Itoa(v.(int))
		case string:
			val = v.(string)
		case []byte:
			val = string(v.([]byte))
		case bool:
			val = strconv.FormatBool(v.(bool))
		default:
			val = ""
		}

		if buffer.Len() > 0 {
			buffer.WriteByte('&')
		}
		buffer.WriteString(key)
		buffer.WriteByte('=')
		buffer.WriteString(url.QueryEscape(val))
	}
	return buffer.String()
}

func format(f string, a ...interface{}) string {
	return fmt.Sprintf(f, a...)
}
