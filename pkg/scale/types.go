package scale

import "time"

// Unit denotes the unit of the weight measurement
type Unit string

const (

	// UnitUnknown denotes an unknown / invalid unit
	UnitUnknown = "--"

	// UnitGrams denotes metric units
	UnitGrams = "g"

	// UnitOz denotes imperial units
	UnitOz = "oz"
)

// State denotes a connection state
type State int

const (

	// StateScanning is active while scanning for a bluetooth device
	StateScanning State = iota

	// StateConnected is active while being connected to the scale
	StateConnected

	// StateDisconnected is active after being disconnected from the scale
	StateDisconnected
)

// ConnectionStatus denotes the current status of the bluetooth device
type ConnectionStatus struct {
	Error error
	State
}

// DataPoint denotes a weight measurement at a certain point in time
type DataPoint struct {
	TimeStamp time.Time
	Unit      Unit
	Weight    float64
}

// Value provides a method to retrieve the current value (for interface use)
func (d DataPoint) Value() float64 {
	return d.Weight
}

// DataPoints denotes a set of data points (usually part of a brew process)
type DataPoints []DataPoint
