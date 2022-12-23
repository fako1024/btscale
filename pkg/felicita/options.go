package felicita

import "github.com/fako1024/gatt"

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
