package protocol

type JsonEnumFields map[string]int
type JsonMsgFields map[string]JsonMsgField

type JsonMsgField struct {
	Rule    string `json:"rule,omitempty"`
	KeyType string `keyType:"rule,omitempty"`
	Type    string `type:"rule,omitempty"`
	Id      string `id:"rule,omitempty"`
}

type JsonMsg struct {
	EnumValues JsonEnumFields `json:"values,omitempty"`
	Fields     JsonMsgFields  `json:"fields,omitempty"`
}

type JsonSchema struct {
	Nested map[string]JsonMsg `json:"nested,omitempty"`
}

