package felicita

import (
	"github.com/fako1024/btscale/pkg/scale"
	"github.com/fako1024/gatt"
)

// BuzzerSetting denotes the basic Buzzer (on touch) setting of the scale
type BuzzerSetting string

const (

	// BuzzerSettingOn denotes that the basic Buzzer (on touch) setting is on
	BuzzerSettingOn = "ON"

	// BuzzerSettingOff denotes that the basic Buzzer (on touch) setting is off
	BuzzerSettingOff = "OFF"
)

// WithDeviceID sets the Bluetooth device ID
func WithDeviceID(deviceID string) func(*Felicita) {
	return func(f *Felicita) {
		f.deviceID = deviceID
	}
}

// WithDeviceName sets the Bluetooth device name
func WithDeviceName(deviceName string) func(*Felicita) {
	return func(f *Felicita) {
		f.deviceName = deviceName
	}
}

// WithDevice sets the Bluetooth device
func WithDevice(btDevice gatt.Device) func(*Felicita) {
	return func(f *Felicita) {
		f.btDevice = btDevice
	}
}

// WithLogger sets a logger
func WithLogger(logger scale.Logger) func(*Felicita) {
	return func(f *Felicita) {
		f.logger = logger
	}
}

// WithForceBuzzerSettingOnConnect ensures that the basic Buzzer (on touch) setting
// is correct upon connection
func WithForceBuzzerSettingOnConnect(setting BuzzerSetting) func(*Felicita) {
	return func(f *Felicita) {
		f.forceBuzzerSettingOnConnect = setting
	}
}
