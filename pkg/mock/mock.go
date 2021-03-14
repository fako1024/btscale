package mock

import (
	"fmt"
	"time"

	"github.com/fako1024/btscale/pkg/scale"
	"github.com/fatih/stopwatch"
)

const (
	defaultDeviceName = "Mock Scale"
	btSettleDelay     = 250 * time.Millisecond
)

// Mock denotes a Mock bluetooth scale
type Mock struct {
	connectionStatus scale.ConnectionStatus
	batteryLevel     byte
	isBuzzingOnTouch bool
	isHighPrecision  bool
	unit             scale.Unit

	timer *stopwatch.Stopwatch

	deviceName string

	stateChangeHandler func(status scale.ConnectionStatus)
	stateChangeChan    chan scale.ConnectionStatus

	dataHandler func(data scale.DataPoint)
	dataChan    chan scale.DataPoint
	doneChan    chan struct{}
}

// New instantiates a new Mock struct
func New() (*Mock, error) {

	// Initialize a new instance of a Mock scale
	f := &Mock{
		deviceName: defaultDeviceName,
		doneChan:   make(chan struct{}),
	}

	return f, f.subscribe()
}

// ConnectionStatus returns the current status of the bluetooth device
func (f *Mock) ConnectionStatus() scale.ConnectionStatus {
	return f.connectionStatus
}

// IsBuzzingOnTouch returns if the scale buzzer is turned on or not (on user interaction)
func (f *Mock) IsBuzzingOnTouch() bool {
	return f.isBuzzingOnTouch
}

// BatteryLevel returns the current battery level
func (f *Mock) BatteryLevel() float64 {
	return float64(f.batteryLevel)
}

// BatteryLevelRaw returns the current battery level in its raw form
func (f *Mock) BatteryLevelRaw() int {
	return int(f.batteryLevel)
}

// Unit returns the current weight unit
func (f *Mock) Unit() scale.Unit {
	return f.unit
}

// SetStateChangeHandler defines a handler function that is called upon state change
func (f *Mock) SetStateChangeHandler(fn func(status scale.ConnectionStatus)) {
	f.stateChangeHandler = fn
}

// SetStateChangeChannel defines a handler function that is called upon state change
func (f *Mock) SetStateChangeChannel(ch chan scale.ConnectionStatus) {
	f.stateChangeChan = ch
}

// SetDataHandler defines a handler function that is called upon retrieval of data
func (f *Mock) SetDataHandler(fn func(data scale.DataPoint)) {
	f.dataHandler = fn
}

// SetDataChannel defines a handler function that is called upon retrieval of data
func (f *Mock) SetDataChannel(ch chan scale.DataPoint) {
	f.dataChan = ch
}

// Tare tares the scale
func (f *Mock) Tare() error {
	return nil
}

// Buzz requests the scale to beep / buzz n times
func (f *Mock) Buzz(n int) (err error) {

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
		time.Sleep(btSettleDelay)
		defer func() {
			if derr := f.ToggleBuzzingOnTouch(); derr != nil {
				err = derr
				return
			}
		}()
	}

	for i := 0; i < n; i++ {

		// Buzz once, then restore former state
		if err = f.ToggleBuzzingOnTouch(); err != nil {
			return
		}
		time.Sleep(btSettleDelay)
		if err = f.ToggleBuzzingOnTouch(); err != nil {
			return
		}
		time.Sleep(btSettleDelay)
	}

	return nil
}

// ToggleBuzzingOnTouch turns the buzzer (on user interaction) on / off
func (f *Mock) ToggleBuzzingOnTouch() error {
	f.isBuzzingOnTouch = !f.isBuzzingOnTouch

	return nil
}

// SetUnit changes the weight unit from / to g / oz
func (f *Mock) SetUnit(unit scale.Unit) error {

	// Check if the unit is already set to the expected value
	if f.unit != scale.UnitUnknown && f.unit == unit {
		return nil
	}

	// Toggle unit, if not
	f.unit = unit
	return nil
}

// TogglePrecision toggles the weight precision between 0.1 and 0.01
func (f *Mock) TogglePrecision() error {
	f.isHighPrecision = !f.isHighPrecision

	return nil
}

// StartTimer starts the timer / stopwatch
func (f *Mock) StartTimer() error {
	if f.timer == nil {
		f.timer = stopwatch.Start(0)
	} else {
		f.timer.Start(0)
	}

	return nil
}

// StopTimer stops the timer / stopwatch
func (f *Mock) StopTimer() error {
	if f.timer != nil {
		f.timer.Stop()
	}

	return nil
}

// ResetTimer resets the timer / stopwatch
func (f *Mock) ResetTimer() error {
	if f.timer != nil {
		f.timer.Reset()
	}

	return nil
}

// ElapsedTime returns the current timer value
func (f *Mock) ElapsedTime() time.Duration {
	if f.timer != nil {
		return f.timer.ElapsedTime()
	}

	return 0
}

// Close terminates the connection to the device
func (f *Mock) Close() error {
	close(f.doneChan)

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (f *Mock) subscribe() error {

	return nil
}
