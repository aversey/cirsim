package cirsim

import (
	"log"
	"math"
)

type Modeler interface {
	conductance(time, delta, voltage, current float64) float64
	current(time, delta, voltage, current float64) float64
	Parameters() map[string]float64
	UpdateParameter(name string, value float64)
}

func newModeler(name string) Modeler {
	switch name {
	case "resistor":
		return newResistor()
	case "capacitor":
		return newCapacitor()
	case "inductor":
		return newInductor()
	case "diode":
		return newDiode()
	case "power":
		return newPower()
	default:
		log.Fatal("wrong component name")
		// compiler wants return here, but it will be never executed:
		return nil
	}
}

type capacitor struct {
	capacitance float64
}

func newCapacitor() *capacitor {
	return &capacitor{0.000001}
}

func (m *capacitor) conductance(time, delta, voltage, current float64) float64 {
	return 2.0 * m.capacitance / delta
}
func (m *capacitor) current(time, delta, voltage, current float64) float64 {
	return -current - voltage*2*m.capacitance/delta
}
func (m *capacitor) Parameters() map[string]float64 {
	return map[string]float64{"Capacitance": m.capacitance}
}
func (m *capacitor) UpdateParameter(name string, value float64) {
	if name == "Capacitance" {
		m.capacitance = value
	}
}

type resistor struct {
	resistance float64
}

func newResistor() *resistor {
	return &resistor{100.0}
}

func (m *resistor) conductance(time, delta, voltage, current float64) float64 {
	return 1.0 / m.resistance
}
func (m *resistor) current(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *resistor) Parameters() map[string]float64 {
	return map[string]float64{"Resistance": m.resistance}
}
func (m *resistor) UpdateParameter(name string, value float64) {
	if name == "Resistance" {
		m.resistance = value
	}
}

type inductor struct {
	inductance float64
}

func newInductor() *inductor {
	return &inductor{0.000001}
}

func (m *inductor) conductance(time, delta, voltage, current float64) float64 {
	return delta / (2.0 * m.inductance)
}
func (m *inductor) current(time, delta, voltage, current float64) float64 {
	return current + voltage*delta/(2*m.inductance)
}
func (m *inductor) Parameters() map[string]float64 {
	return map[string]float64{"Inductance": m.inductance}
}
func (m *inductor) UpdateParameter(name string, value float64) {
	if name == "Inductance" {
		m.inductance = value
	}
}

type diode struct{}

func newDiode() *diode {
	return &diode{}
}

func (m *diode) conductance(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *diode) current(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *diode) Parameters() map[string]float64 {
	return map[string]float64{}
}
func (m *diode) UpdateParameter(name string, value float64) {}

type power struct {
	maxCurrent float64
	frequency  float64
}

func newPower() *power {
	return &power{maxCurrent: 1.0, frequency: 1000.0}
}

func (m *power) conductance(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *power) current(time, delta, voltage, current float64) float64 {
	return m.maxCurrent * math.Sin(time*m.frequency*2*math.Pi)
}
func (m *power) Parameters() map[string]float64 {
	return map[string]float64{"Current": m.maxCurrent, "Frequency": m.frequency}
}
func (m *power) UpdateParameter(name string, value float64) {
	if name == "Current" {
		m.maxCurrent = value
	} else if name == "Frequency" {
		m.frequency = value
	}
}
