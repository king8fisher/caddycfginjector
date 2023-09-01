package util

import (
	"encoding/json"
	"strings"
)

type IDField struct {
	Id string `json:"@id"`
}

// encodeAtId returns "@id":"<id>" encoded with
// proper character escaping for the id field.
func encodeAtId(id string) string {
	var s strings.Builder
	e := json.NewEncoder(&s)
	err := e.Encode(IDField{Id: id})
	if err != nil {
		return ""
	}
	stripped := strings.TrimSuffix(s.String(), "}\n")
	stripped = strings.TrimPrefix(stripped, "{")
	return stripped
}

// EncodeJSONString returns "value" properly escaped for JSON and surrounded with quotes.
func EncodeJSONString(value string) string {
	return strings.TrimPrefix(encodeAtId(value), `"@id":`)
}
