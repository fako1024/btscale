# go package to facilitate data access and management of bluetooth-based scales
This package allows to extract structured data from bluetooth-based remote scales and provides a management interface to control said devices. Usage is fairly trivial (see examples directory for a simple console logger implementation and a tool for controlling basic functions).

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

## Example
```go
	// Initialize a new Felicita bluetooth scale
	s, err := felicita.New()
	if err != nil {
		logrus.StandardLogger().Fatalf("Error opening Felicita scale: %s", err)
	}

	// Set a data channel to continuously log incoming data
	dataChan := make(chan *scale.DataPoint, 256)
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
For additional examples please refer to the `examples` folder.
