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

	minBatteryLevel = 129.
	maxBatteryLevel = 158.

	cmdStartTimer = 0x52
	cmdStopTimer  = 0x53
	cmdResetTimer = 0x43

	cmdToggleBuzzer    = 0x42
	cmdTogglePrecision = 0x44
	cmdTare            = 0x54
	cmdToggleUnit      = 0x55

	btSettleDelay   = 50 * time.Millisecond
	btSettleRetries = 100
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
	batteryLevel     byte
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
	return parseBatteryLevel(f.batteryLevel)
}

// BatteryLevelRaw returns the current battery level in its raw form
func (f *Felicita) BatteryLevelRaw() int {
	return int(f.batteryLevel)
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
func (f *Felicita) Buzz(n int) (err error) {

	if n <= 0 {
		return fmt.Errorf("invalid number of beeps requested: %d", n)
	}

	// If the buzzer is currently turned on, shortly turn it off and ensure it is
	// re-enabled at the end of the function. In this case, n is reduced by one since
	// enabling the buzzer will cause yet another buzz at the end
	if f.IsBuzzingOnTouch() {
		if err = f.ToggleBuzzingOnTouch(); err != nil {
			return
		}
		n--
		if err = f.waitForBuzzer(false); err != nil {
			return
		}

		defer func() {
			if derr := f.ToggleBuzzingOnTouch(); derr != nil {
				err = derr
				return
			}
			if derr := f.waitForBuzzer(true); derr != nil {
				err = derr
				return
			}
		}()
	}

	for i := 0; i < n; i++ {

		// Buzz once, then restore former state
		if err = f.buzzAndRestore(); err != nil {
			return
		}
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
		return fmt.Errorf("failed to write to uninitialized device")
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
		if !strings.EqualFold(p.Name(), f.deviceName) {
			return
		}
		if f.deviceID != "" {

			if !strings.EqualFold(p.ID(), f.deviceID) {
				return
			}
		}

		// Stop scanning once we've got the peripheral we're looking for.
		p.Device().StopScanning()
		p.Device().Connect(p)
	}
}

func (f *Felicita) onPeriphConnected(p gatt.Peripheral, connErr error) {

	f.setStatus(scale.StateConnected, nil)
	defer func() {
		p.Device().CancelConnection(p)
		f.setStatus(scale.StateDisconnected, connErr)
	}()

	// Set connection MTU
	if err := p.SetMTU(500); err != nil {
		connErr = fmt.Errorf("failed to set MTU: %s", err)
		return
	}

	// Discover services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		connErr = fmt.Errorf("failed to discover services: %s", err)
		return
	}
	for _, s := range ss {
		if s.UUID().String() == dataService {

			// Discover characteristics
			cs, err := p.DiscoverCharacteristics(nil, s)
			if err != nil {
				connErr = fmt.Errorf("failed to discover characteristics: %s", err)
				return
			}
			for _, c := range cs {
				if c.UUID().String() == dataCharacteristic {
					f.btPeripheral = p
					f.btCharacteristic = c

					// Discover descriptors
					_, err := p.DiscoverDescriptors(nil, c)
					if err != nil {
						connErr = fmt.Errorf("failed to discover descriptors: %s", err)
						return
					}

					if err := p.SetNotifyValue(c, f.receiveData); err != nil {
						connErr = fmt.Errorf("failed to subscribe characteristic: %s", err)
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

	// Cannot explicitly handle / return error here, but an error should be
	// raised elsewhere anyways
	_ = f.btDevice.Init(f.onStateChanged)
}

func (f *Felicita) receiveData(c *gatt.Characteristic, req []byte, err error) {

	if err != nil || len(req) != 18 {
		return
	}

	weight, convErr := strconv.ParseFloat(string(req[2:9]), 64)
	if convErr != nil {
		return
	}
	dataPoint := scale.DataPoint{
		TimeStamp: time.Now(),
		Weight:    weight / 100.,
		Unit:      parseUnit(req[9:11]),
	}
	f.batteryLevel = req[15]
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

func (f *Felicita) buzzAndRestore() (err error) {
	if err = f.ToggleBuzzingOnTouch(); err != nil {
		return
	}
	if err = f.waitForBuzzer(true); err != nil {
		return
	}
	if err = f.ToggleBuzzingOnTouch(); err != nil {
		return
	}
	if err = f.waitForBuzzer(false); err != nil {
		return
	}

	return
}

func (f *Felicita) waitForBuzzer(targetState bool) error {
	for i := 0; i < btSettleRetries; i++ {
		if f.IsBuzzingOnTouch() == targetState {
			return nil
		}
		time.Sleep(btSettleDelay)
	}

	return fmt.Errorf("target buzzer state %v was not reached within %v", targetState, time.Duration(btSettleRetries)*btSettleDelay)
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
	return data == 0x22
}
