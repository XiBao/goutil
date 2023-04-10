package goutil

import "encoding/json"

func JSONMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
	// buf := GetBufferPool()
	// defer PutBufferPool(buf)
	// encoder := json.NewEncoder(buf)
	// if err := encoder.Encode(v); err != nil {
	// 	return nil, err
	// }
	// bs := buf.Bytes()
	// return bs[:buf.Len()-1], nil
}

func JSONUnmarshal(bs []byte, v any) error {
	buf := GetBufferPool()
	defer PutBufferPool(buf)
	buf.Write(bs)
	return json.NewDecoder(buf).Decode(v)
}
