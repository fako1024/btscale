package felicita

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
