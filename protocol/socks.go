package protocol

import (
	"singboxconfig/entity"
)

type SocksNode struct {
	Tag        string `json:"tag"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
}

func DecodeSocksURL(socksURL string) (*SocksNode, error) {
	return nil, nil
}

func ConvertSocksToSingBox(item *SocksNode) (*entity.SingBoxOut, error) {
	return nil, nil
}
