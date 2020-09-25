# go package to facilitate data access and management of bluetooth-based scales
This package allows to extract structured data from bluetooth-based remote scales and provides a management interface to control said devices. Usage is fairly trivial (see examples directory for a simple console logger implementation and a tool for controlling basic functions).

**NOTE: This package is currently work in progress. Interfaces and implementation are subject to change.**

## Features
- Control of basic settings (via multiple interfaces)
  - Status
  - Battery level
  - Weight unit
  - Measurement precision
  - Buzzer
- Extraction of scale data (both channel and handler concept supported)  
	- Timestamp
	- Weight / Unit
- Timer functionality

## Installation
```bash
go get -u github.com/fako1024/btscale
```

## API summary
The following API interfaces / methods are exposed:
```
// Basic denotes a basic coffee scale
type Basic interface {

	// ConnectionStatus returns the current connection status of the scale device
	ConnectionStatus() ConnectionStatus

	// BatteryLevel returns the current battery level
	BatteryLevel() float64

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
```

## Example
```go
	// Initialize a new Felicita bluetooth scale
	s, err := felicita.New()
	if err != nil {
		logrus.StandardLogger().Fatalf("Error opening Felicita scale: %s", err)
	}

	// Set a data channel to continuously log incoming data
	dataChan := make(chan scale.DataPoint, 256)
	s.SetDataChannel(dataChan)

	// Setup a state handler to notify upon connection status change
	s.SetStateChangeHandler(func(status scale.ConnectionStatus) {
		log.Warnf("State change: %v", status)
	})

	// Setup a signal channel to gracefully disconnect the bluetooth device upon termination
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Infof("Got signal, terminating connection to device")
		s.Close()
		os.Exit(0)
	}()

	// Continuously read from channel
	for {
		select {
		default:
			log.Warnf("Read DATA from Channel: %v, %v, %v, %v", *(<-dataChan), s.ConnectionStatus(), s.BatteryLevel(), s.IsBuzzingOnTouch())
		}
	}
}()
```
For additional examples please refer to the `cmd` folder.
