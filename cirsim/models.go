package cirsim

import "math"

type Modeler interface {
	ModelConductance(time, delta, voltage, current float64) float64
	ModelCurrent(time, delta, voltage, current float64) float64
	Parameters() map[string]float64
	UpdateParameter(name string, value float64)
}

type Capacitor struct {
	Capacitance float64
}

func (m *Capacitor) ModelConductance(time, delta, voltage, current float64) float64 {
	return 2.0 * m.Capacitance / delta
}
func (m *Capacitor) ModelCurrent(time, delta, voltage, current float64) float64 {
	return -current - voltage*2*m.Capacitance/delta
}
func (m *Capacitor) Parameters() map[string]float64 {
	return map[string]float64{"Capacitance": m.Capacitance}
}
func (m *Capacitor) UpdateParameter(name string, value float64) {
	if name == "Capacitance" {
		m.Capacitance = value
	}
}

type Resistor struct {
	Resistance float64
}

func (m *Resistor) ModelConductance(time, delta, voltage, current float64) float64 {
	return 1.0 / m.Resistance
}
func (m *Resistor) ModelCurrent(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *Resistor) Parameters() map[string]float64 {
	return map[string]float64{"Resistance": m.Resistance}
}
func (m *Resistor) UpdateParameter(name string, value float64) {
	if name == "Resistance" {
		m.Resistance = value
	}
}

type Inductor struct {
	Inductance float64
}

func (m *Inductor) ModelConductance(time, delta, voltage, current float64) float64 {
	return delta / (2.0 * m.Inductance)
}
func (m *Inductor) ModelCurrent(time, delta, voltage, current float64) float64 {
	return current + voltage*delta/(2*m.Inductance)
}
func (m *Inductor) Parameters() map[string]float64 {
	return map[string]float64{"Inductance": m.Inductance}
}
func (m *Inductor) UpdateParameter(name string, value float64) {
	if name == "Inductance" {
		m.Inductance = value
	}
}

type Diode struct{}

func (m *Diode) ModelConductance(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *Diode) ModelCurrent(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *Diode) Parameters() map[string]float64 {
	return map[string]float64{}
}
func (m *Diode) UpdateParameter(name string, value float64) {}

type Power struct {
	Current   float64
	Frequency float64
}

func (m *Power) ModelConductance(time, delta, voltage, current float64) float64 {
	return 0
}
func (m *Power) ModelCurrent(time, delta, voltage, current float64) float64 {
	return m.Current * math.Sin(time*m.Frequency*2*math.Pi)
}
func (m *Power) Parameters() map[string]float64 {
	return map[string]float64{"Current": m.Current, "Frequency": m.Frequency}
}
func (m *Power) UpdateParameter(name string, value float64) {
	if name == "Current" {
		m.Current = value
	} else if name == "Frequency" {
		m.Frequency = value
	}
}
