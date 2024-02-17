package jsontransfer

import "encoding/json"

func GetJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return data, nil
}
