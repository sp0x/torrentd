package json

import "encoding/json"

const name = "json"

//Serializer is a json serializer.
var Serializer = new(jsonSerializer)

type jsonSerializer int

func (j jsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonSerializer) Unmarshal(b []byte, output interface{}) error {
	return json.Unmarshal(b, output)
}

func (j jsonSerializer) Name() string {
	return name
}
