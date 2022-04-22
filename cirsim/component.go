package cirsim

type ComponentSettings interface {
	Nodes() [2]int
	ModelName() string
}

type component struct {
	Modeler
	currentOverTime []float64
	nodes           [2]int
}

func newComponent(settings ComponentSettings) *component {
	var c component
	c.nodes = settings.Nodes()
	c.Modeler = newModeler(settings.ModelName())
	c.currentOverTime = make([]float64, iterations)
	for i := range c.currentOverTime {
		c.currentOverTime[i] = 0
	}
	return &c
}
