package felicita

import "github.com/fako1024/gatt"

var (
	defaultBTClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, true),
	}
)
