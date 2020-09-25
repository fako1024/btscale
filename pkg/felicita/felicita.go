package felicita

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/fako1024/btscale/pkg/scale"
	"github.com/fako1024/gatt"
	"github.com/fatih/stopwatch"
)

const (
	defaultDeviceName  = "FELICITA"
	dataService        = "ffe0"
	dataCharacteristic = "ffe1"

	minBatteryLevel = 137.
	maxBatteryLevel = 158.

	cmdStartTimer = 0x52
	cmdStopTimer  = 0x53
	cmdResetTimer = 0x43

	cmdToggleBuzzer    = 0x42
	cmdTogglePrecision = 0x44
	cmdTare            = 0x54
	cmdToggleUnit      = 0x55

	btSettleDelay = 250 * time.Millisecond
)

var (
	defaultBTClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, true),
	}
)

// Felicita denotes a Felicita bluetooth scale
type Felicita struct {
	connectionStatus scale.ConnectionStatus
	batteryLevel     float64
	isBuzzingOnTouch bool
	unit             scale.Unit

	timer *stopwatch.Stopwatch

	deviceID   string
	deviceName string

	stateChangeHandler func(status scale.ConnectionStatus)
	stateChangeChan    chan scale.ConnectionStatus

	dataHandler func(data scale.DataPoint)
	dataChan    chan scale.DataPoint
	doneChan    chan struct{}

	btDevice         gatt.Device
	btPeripheral     gatt.Peripheral
	btCharacteristic *gatt.Characteristic
}

// New instantiates a new Felicita struct, executing functional options, if any
func New(options ...func(*Felicita)) (*Felicita, error) {

	// Initialize a new instance of a Felicita scale
	f := &Felicita{
		deviceName: defaultDeviceName,
		doneChan:   make(chan struct{}),
	}

	// Execute functional options (if any), see options.go for implementation
	for _, option := range options {
		option(f)
	}

	// Initialize a new GATT device
	btDevice, err := gatt.NewDevice(defaultBTClientOptions...)
	if err != nil {
		return nil, err
	}
	f.btDevice = btDevice

	return f, f.subscribe()
}

// ConnectionStatus returns the current status of the bluetooth device
func (f *Felicita) ConnectionStatus() scale.ConnectionStatus {
	return f.connectionStatus
}

// IsBuzzingOnTouch returns if the scale buzzer is turned on or not (on user interaction)
func (f *Felicita) IsBuzzingOnTouch() bool {
	return f.isBuzzingOnTouch
}

// BatteryLevel returns the current battery level
func (f *Felicita) BatteryLevel() float64 {
	return f.batteryLevel
}

// Unit returns the current weight unit
func (f *Felicita) Unit() scale.Unit {
	return f.unit
}

// SetStateChangeHandler defines a handler function that is called upon state change
func (f *Felicita) SetStateChangeHandler(fn func(status scale.ConnectionStatus)) {
	f.stateChangeHandler = fn
}

// SetStateChangeChannel defines a handler function that is called upon state change
func (f *Felicita) SetStateChangeChannel(ch chan scale.ConnectionStatus) {
	f.stateChangeChan = ch
}

// SetDataHandler defines a handler function that is called upon retrieval of data
func (f *Felicita) SetDataHandler(fn func(data scale.DataPoint)) {
	f.dataHandler = fn
}

// SetDataChannel defines a handler function that is called upon retrieval of data
func (f *Felicita) SetDataChannel(ch chan scale.DataPoint) {
	f.dataChan = ch
}

// Tare tares the scale
func (f *Felicita) Tare() error {
	return f.write(cmdTare)
}

// Buzz requests the scale to beep / buzz n times
func (f *Felicita) Buzz(n int) error {

	if n <= 0 {
		return fmt.Errorf("Invalid number of beeps requested: %d", n)
	}

	// If the buzzer is currently turned on, shortly turn it off and ensure it is
	// re-enabled at the end of the function. In this case, n is reduced by one since
	// enabling the buzzer will cause yet another buzz at the end
	if f.IsBuzzingOnTouch() {
		if err := f.ToggleBuzzingOnTouch(); err != nil {
			return err
		}
		n--
		time.Sleep(btSettleDelay)
		defer f.ToggleBuzzingOnTouch()
	}

	for i := 0; i < n; i++ {

		// Buzz once, then restore former state
		if err := f.ToggleBuzzingOnTouch(); err != nil {
			return err
		}
		time.Sleep(btSettleDelay)
		if err := f.ToggleBuzzingOnTouch(); err != nil {
			return err
		}
		time.Sleep(btSettleDelay)
	}

	return nil
}

// ToggleBuzzingOnTouch turns the buzzer (on user interaction) on / off
func (f *Felicita) ToggleBuzzingOnTouch() error {
	return f.write(cmdToggleBuzzer)
}

// SetUnit changes the weight unit from / to g / oz
func (f *Felicita) SetUnit(unit scale.Unit) error {

	// Check if the unit is already set to the expected value
	if f.unit != scale.UnitUnknown && f.unit == unit {
		return nil
	}

	// Toggle unit, if not
	return f.write(cmdToggleUnit)
}

// TogglePrecision toggles the weight precision between 0.1 and 0.01
func (f *Felicita) TogglePrecision() error {
	return f.write(cmdTogglePrecision)
}

// StartTimer starts the timer / stopwatch
func (f *Felicita) StartTimer() error {
	if err := f.write(cmdStartTimer); err != nil {
		return err
	}

	if f.timer == nil {
		f.timer = stopwatch.Start(0)
	} else {
		f.timer.Start(0)
	}

	return nil
}

// StopTimer stops the timer / stopwatch
func (f *Felicita) StopTimer() error {
	if err := f.write(cmdStopTimer); err != nil {
		return err
	}

	if f.timer != nil {
		f.timer.Stop()
	}

	return nil
}

// ResetTimer resets the timer / stopwatch
func (f *Felicita) ResetTimer() error {
	if err := f.write(cmdResetTimer); err != nil {
		return err
	}

	if f.timer != nil {
		f.timer.Reset()
	}

	return nil
}

// ElapsedTime returns the current timer value
func (f *Felicita) ElapsedTime() time.Duration {
	if f.timer != nil {
		return f.timer.ElapsedTime()
	}

	return 0
}

// Close terminates the connection to the device
func (f *Felicita) Close() error {
	close(f.doneChan)

	return f.btDevice.RemoveAllServices()
}

////////////////////////////////////////////////////////////////////////////////

func (f *Felicita) subscribe() error {

	// Register handlers
	f.btDevice.Handle(
		gatt.PeripheralDiscovered(f.genOnPeriphDiscovered()),
		gatt.PeripheralConnected(f.onPeriphConnected),
		gatt.PeripheralDisconnected(f.onPeriphDisconnected),
	)

	// Initialize the device
	return f.btDevice.Init(f.onStateChanged)
}

func (f *Felicita) setStatus(state scale.State, err error) {
	f.connectionStatus = scale.ConnectionStatus{
		State: state,
		Error: err,
	}

	// Call handler function, if any
	if f.stateChangeHandler != nil {
		f.stateChangeHandler(f.connectionStatus)
	}

	// Put data point on channel, if any
	if f.stateChangeChan != nil {
		f.stateChangeChan <- f.connectionStatus
	}
}

func (f *Felicita) write(cmd byte) error {
	if f.btPeripheral == nil || f.btCharacteristic == nil {
		return fmt.Errorf("Failed to write to uninitialized device")
	}

	return f.btPeripheral.WriteCharacteristic(f.btCharacteristic, []byte{cmd}, false)
}

////////////////////////////////////////////////////////////////////////////////

func (f *Felicita) onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		f.setStatus(scale.StateScanning, nil)
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (f *Felicita) genOnPeriphDiscovered() func(p gatt.Peripheral, arg2 *gatt.Advertisement, arg3 int) {
	return func(p gatt.Peripheral, arg2 *gatt.Advertisement, arg3 int) {

		// Check if name and / or device ID have been overridden
		if strings.ToUpper(p.Name()) != strings.ToUpper(f.deviceName) {
			return
		}
		if f.deviceID != "" {
			if strings.ToUpper(p.ID()) != strings.ToUpper(f.deviceID) {
				return
			}
		}

		// Stop scanning once we've got the peripheral we're looking for.
		p.Device().StopScanning()
		p.Device().Connect(p)
	}
}

func (f *Felicita) onPeriphConnected(p gatt.Peripheral, err error) {

	f.setStatus(scale.StateConnected, nil)
	var connErr error
	defer func() {
		p.Device().CancelConnection(p)
		f.setStatus(scale.StateDisconnected, connErr)
	}()

	// Set connection MTU
	if err := p.SetMTU(500); err != nil {
		connErr = fmt.Errorf("Failed to set MTU: %s", err)
		return
	}

	// Discover services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		connErr = fmt.Errorf("Failed to discover services: %s", err)
		return
	}
	for _, s := range ss {
		if s.UUID().String() == dataService {

			// Discover characteristics
			cs, err := p.DiscoverCharacteristics(nil, s)
			if err != nil {
				connErr = fmt.Errorf("Failed to discover characteristics: %s", err)
				return
			}
			for _, c := range cs {
				if c.UUID().String() == dataCharacteristic {
					f.btPeripheral = p
					f.btCharacteristic = c

					// Discover descriptors
					_, err := p.DiscoverDescriptors(nil, c)
					if err != nil {
						connErr = fmt.Errorf("Failed to discover descriptors: %s", err)
						return
					}

					if err := p.SetNotifyValue(c, f.receiveData); err != nil {
						connErr = fmt.Errorf("Failed to subscribe characteristic: %s", err)
						return
					}
				}
			}
		}
	}

	<-f.doneChan
}

func (f *Felicita) onPeriphDisconnected(p gatt.Peripheral, err error) {
	close(f.doneChan)
	f.doneChan = make(chan struct{})

	f.btDevice.Init(f.onStateChanged)
}

func (f *Felicita) receiveData(c *gatt.Characteristic, req []byte, err error) {

	if len(req) != 18 {
		return
	}

	weight, err := strconv.ParseFloat(string(req[2:9]), 64)
	if err != nil {
		return
	}
	dataPoint := scale.DataPoint{
		TimeStamp: time.Now(),
		Weight:    weight / 100.,
		Unit:      parseUnit(req[9:11]),
	}
	f.batteryLevel = parseBatteryLevel(req[15])
	f.isBuzzingOnTouch = parseSignalFlag(req[14])
	f.unit = dataPoint.Unit

	// Call handler function, if any
	if f.dataHandler != nil {
		f.dataHandler(dataPoint)
	}

	// Put data point on channel, if any
	if f.dataChan != nil {
		f.dataChan <- dataPoint
	}
}

////////////////////////////////////////////////////////////////////////////////

func parseUnit(data []byte) scale.Unit {
	if len(data) != 2 {
		return scale.UnitUnknown
	}

	if strings.Contains(strings.ToLower(string(data)), "g") {
		return scale.UnitGrams
	}
	if strings.Contains(strings.ToLower(string(data)), "oz") {
		return scale.UnitOz
	}

	return scale.UnitUnknown
}

func parseBatteryLevel(data byte) float64 {

	val := int(data)
	if val < minBatteryLevel {
		return 0.
	} else if val > maxBatteryLevel {
		return 1.
	}

	return math.Round((float64(val)-minBatteryLevel)/(maxBatteryLevel-minBatteryLevel)*100.) / 100.
}

func parseSignalFlag(data byte) bool {
	if data == 0x22 {
		return true
	}

	return false
}
