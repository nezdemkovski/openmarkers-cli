package models

import (
	"encoding/json"
	"strconv"
)

type NumValue struct {
	Num *float64
	Str string
}

func (n *NumValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		n.Num = &f
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			n.Num = &f
		} else {
			n.Str = s
		}
		return nil
	}

	return &json.UnmarshalTypeError{Value: string(data), Type: nil}
}

func (n NumValue) MarshalJSON() ([]byte, error) {
	if n.Num != nil {
		return json.Marshal(*n.Num)
	}
	if n.Str != "" {
		return json.Marshal(n.Str)
	}
	return []byte("null"), nil
}
