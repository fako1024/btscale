package scale

import "time"

// Basic denotes a basic coffee scale
type Basic interface {

	// ConnectionStatus returns the current connection status of the scale device
	ConnectionStatus() ConnectionStatus

	// BatteryLevel returns the current battery level
	BatteryLevel() float64

	// BatteryLevelRaw returns the current battery level in its raw form
	BatteryLevelRaw() int

	// Unit returns the current weight unit
	Unit() Unit

	// SetUnit sets the weight unit
	SetUnit(unit Unit) error

	// Tare tares the scale
	Tare() error

	// TogglePrecision toggles the weight precision between 0.1 and 0.01
	TogglePrecision() error

	// SetStateChangeHandler defines a handler function that is called upon state change
	SetStateChangeHandler(fn func(status ConnectionStatus))

	// SetStateChangeChannel defines a handler function that is called upon state change
	SetStateChangeChannel(ch chan ConnectionStatus)

	// SetDataHandler defines a handler function that is called upon retrieval of data
	SetDataHandler(fn func(data DataPoint))

	// SetDataChannel defines a handler function that is called upon retrieval of data
	SetDataChannel(ch chan DataPoint)

	// Close terminates the connection to the device
	Close() error
}

// Buzzer denotes audible signaling functionality
type Buzzer interface {

	// IsBuzzingOnTouch returs if the buzzer  (on user interaction) is on / off
	IsBuzzingOnTouch() bool

	// ToggleBuzzingOnTouch turns the buzzer (on user interaction) on / off
	ToggleBuzzingOnTouch() error

	// Buzz requests the scale to beep / buzz n times
	Buzz(n int) error
}

// Timer denotes timer / stopwatch functionality
type Timer interface {

	// StartTimer starts the timer / stopwatch
	StartTimer() error

	// StopTimer stops the timer / stopwatch
	StopTimer() error

	// ResetTimer resets the timer / stopwatch
	ResetTimer() error

	// ElapsedTime returns the current timer value
	ElapsedTime() time.Duration
}

// WithTimer denotes a scale with timer functionality
type WithTimer interface {
	Basic
	Timer
}

// WithBuzzer denotes a scale with buzzer functionality
type WithBuzzer interface {
	Basic
	Buzzer
}

// Scale denotes the "default" scale containing all functionality
type Scale interface {
	Basic
	Buzzer
	Timer
}
