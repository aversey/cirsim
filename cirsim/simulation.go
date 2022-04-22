package cirsim

import "gonum.org/v1/gonum/mat"

const (
	defaultPeriod float64 = 0.01
	iterations    int     = 1000
)

type Simulator interface {
	Period() float64
	SetPeriod(float64)
	VoltageRange() (float64, float64)
	CurrentRange() (float64, float64)
	VoltagesOfNode(i int) []float64
	CurrentsOfComponent(i int) []float64
	ModelerOfComponent(i int) Modeler
	Simulate()
}

type simulation struct {
	period       float64
	voltageMax   float64
	voltageMin   float64
	currentMax   float64
	currentMin   float64
	nodeVoltages [][]float64
	components   []*component
}

func New(nodesCount int, components []ComponentSettings) Simulator {
	var sim simulation
	sim.period = defaultPeriod
	sim.nodeVoltages = make([][]float64, nodesCount)
	for i := range sim.nodeVoltages {
		sim.nodeVoltages[i] = make([]float64, iterations)
	}
	sim.components = make([]*component, 0)
	for _, c := range components {
		sim.components = append(sim.components, newComponent(c))
	}
	sim.Simulate()
	return &sim
}

func (sim *simulation) Period() float64 {
	return sim.period
}
func (sim *simulation) SetPeriod(period float64) {
	sim.period = period
}
func (sim *simulation) VoltageRange() (float64, float64) {
	return sim.voltageMin, sim.voltageMax
}
func (sim *simulation) CurrentRange() (float64, float64) {
	return sim.currentMin, sim.currentMax
}
func (sim *simulation) VoltagesOfNode(i int) []float64 {
	return sim.nodeVoltages[i]
}
func (sim *simulation) CurrentsOfComponent(i int) []float64 {
	return sim.components[i].currentOverTime
}
func (sim *simulation) ModelerOfComponent(i int) Modeler {
	return sim.components[i]
}

func (sim *simulation) Simulate() {
	sim.nullify()
	N := len(sim.nodeVoltages)
	delta := sim.period / float64(iterations)
	for i := 0; i != iterations; i++ {
		// fill conductances and currents:
		time := float64(i) * sim.period / float64(iterations)
		conductances := mat.NewDense(N+1, N, nil)
		currents := mat.NewVecDense(N+1, nil)
		for _, c := range sim.components {
			voltage := sim.nodeVoltages[c.nodes[1]][i] -
				sim.nodeVoltages[c.nodes[0]][i]
			current := c.current(time, delta, voltage, c.currentOverTime[i])
			cond := c.conductance(time, delta, voltage, c.currentOverTime[i])
			currents.SetVec(c.nodes[1], currents.AtVec(c.nodes[1])-current)
			currents.SetVec(c.nodes[0], currents.AtVec(c.nodes[0])+current)
			conductances.Set(c.nodes[1], c.nodes[1],
				conductances.At(c.nodes[1], c.nodes[1])+cond)
			conductances.Set(c.nodes[0], c.nodes[0],
				conductances.At(c.nodes[0], c.nodes[0])+cond)
			conductances.Set(c.nodes[1], c.nodes[0],
				conductances.At(c.nodes[1], c.nodes[0])-cond)
			conductances.Set(c.nodes[0], c.nodes[1],
				conductances.At(c.nodes[0], c.nodes[1])-cond)
		}
		// set up additional row to determine ground node:
		groundNodeRow := make([]float64, N)
		groundNodeRow[0] = 1
		for j := 1; j != N; j++ {
			groundNodeRow[j] = 0
		}
		conductances.SetRow(N, groundNodeRow)
		currents.SetVec(N, 0)
		// solve the equation:
		voltages := mat.NewVecDense(N, nil)
		voltages.SolveVec(conductances, currents)
		// save results:
		for j, n := range sim.nodeVoltages {
			n[i] = voltages.AtVec(j)
		}
		for _, c := range sim.components {
			voltage := voltages.AtVec(c.nodes[1]) - voltages.AtVec(c.nodes[0])
			current := c.current(time, delta, voltage, c.currentOverTime[i])
			cond := c.conductance(time, delta, voltage, c.currentOverTime[i])
			c.currentOverTime[i] = voltage*cond + current
		}
	}
	sim.updateRanges()
}

func (sim *simulation) nullify() {
	for _, n := range sim.nodeVoltages {
		for i := range n {
			n[i] = 0
		}
	}
	for _, c := range sim.components {
		for i := range c.currentOverTime {
			c.currentOverTime[i] = 0
		}
	}
}

func (sim *simulation) updateRanges() {
	sim.voltageMax = 0
	sim.voltageMin = 0
	for _, n := range sim.nodeVoltages {
		for _, v := range n {
			if v > sim.voltageMax {
				sim.voltageMax = v
			} else if v < sim.voltageMin {
				sim.voltageMin = v
			}
		}
	}
	sim.currentMax = sim.components[0].currentOverTime[0]
	sim.currentMin = sim.components[0].currentOverTime[0]
	for _, comp := range sim.components {
		for _, c := range comp.currentOverTime {
			if c > sim.currentMax {
				sim.currentMax = c
			} else if c < sim.currentMin {
				sim.currentMin = c
			}
		}
	}
}
