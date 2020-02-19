package serializers

import(
    "encoding/json"
)
type DataDumperSerializer struct {

}

func DataDumper(config map[string]interface{}) *DataDumperSerializer {
    return &DataDumperSerializer{}
}

func (s *DataDumperSerializer) Thaw(data []byte) (decoded interface{}, err error) {
    err = json.Unmarshal(data, &decoded)
    return
}

func (s *DataDumperSerializer) Freeze(data interface{}) (encoded []byte, err error) {
    return json.Marshal(data)
}


