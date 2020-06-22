package models

import (
	"encoding/json"
)

type Sponsor struct {
	Name string `json:"Name"`
	Id string `json:"Id"`
	Color string `json:"Color"`
}

func (s *Sponsor) Init() *Sponsor {
	return s
}

func (s Sponsor) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s Sponsor) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}
