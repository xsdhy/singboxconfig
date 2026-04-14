package singbox

import (
	"singboxconfig/entity"
)

func GetExperimental(device string) *entity.SingExperimental {
	if device == entity.DevicePhone {
		return nil
	}

	addr := "127.0.0.1:9090"
	if device == entity.DeviceTV {
		addr = "192.168.10.66:9090"
	}

	return &entity.SingExperimental{
		CacheFile: entity.SingCacheFile{
			Enabled: true,
		},
		ClashAPI: entity.SingClashAPI{
			ExternalController:       addr,
			ExternalUI:               "ui",
			ExternalUIDownloadURL:    "https://file.xsdhy.com/files/Yacd-meta-gh-pages.zip",
			ExternalUIDownloadDetour: "direct",
			DefaultMode:              "rule",
		},
	}
}
