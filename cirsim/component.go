package cirsim

import (
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type Component struct {
	widget.BaseWidget
	Pos             fyne.Position
	Model           Modeler
	CurrentOverTime []float64
	ANode           *Node
	BNode           *Node
	CurrentRange    chart.Range
	Image           *canvas.Image
	labels          []*widget.Label
	entries         []*widget.Entry
}

func NewComponent(settings io.Reader, r chart.Range, nodes []*Node, update func()) *Component {
	var c Component
	var t string
	var a int
	var b int
	_, err := fmt.Fscanf(settings, "%f %f %s %d %d\n",
		&c.Pos.X, &c.Pos.Y, &t, &a, &b,
	)
	if err != nil {
		return nil
	}
	if a <= 0 || b <= 0 {
		log.Fatal("unconnected components in the circuit")
	}
	c.ANode = nodes[a-1]
	nodes[a-1].AComponents = append(nodes[a-1].AComponents, &c)
	c.BNode = nodes[b-1]
	nodes[b-1].BComponents = append(nodes[b-1].BComponents, &c)
	switch t {
	case "resistor":
		c.Model = &Resistor{100.0}
	case "capacitor":
		c.Model = &Capacitor{0.000001}
	case "inductor":
		c.Model = &Inductor{0.000001}
	case "diode":
		c.Model = &Diode{}
	case "power":
		c.Model = &Power{Current: 1.0, Frequency: 1000.0}
	}
	c.entries = make([]*widget.Entry, 0)
	c.labels = make([]*widget.Label, 0)
	for k, v := range c.Model.Parameters() {
		e := widget.NewEntry()
		e.TextStyle.Monospace = true
		e.SetPlaceHolder(k)
		e.SetText(fmt.Sprintf("%f", v))
		e.OnSubmitted = func(s string) {
			var v float64
			_, err := fmt.Sscanf(s+"\n", "%f\n", &v)
			if err != nil {
				e.SetText(fmt.Sprintf("%f", c.Model.Parameters()[k]))
			} else {
				c.Model.UpdateParameter(k, v)
				update()
			}
		}
		c.entries = append(c.entries, e)
		l := widget.NewLabel(k)
		l.TextStyle.Monospace = true
		c.labels = append(c.labels, l)
	}
	c.CurrentRange = r
	c.CurrentOverTime = make([]float64, 1000)
	for i := range c.CurrentOverTime {
		c.CurrentOverTime[i] = 0
	}
	c.renderChart()
	return &c
}

func (c *Component) renderChart() {
	graph := chart.Chart{
		Width:        140,
		Height:       60,
		ColorPalette: &componentColorPalette{},
		XAxis:        chart.HideXAxis(),
		YAxis: chart.YAxis{
			Style: chart.Hidden(),
			Range: c.CurrentRange,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: chart.LinearRange(0, 999),
				YValues: c.CurrentOverTime,
			},
		},
	}
	writer := &chart.ImageWriter{}
	graph.Render(chart.PNG, writer)
	img, _ := writer.Image()
	c.Image = canvas.NewImageFromImage(img)
}

type componentColorPalette struct{}

func (*componentColorPalette) BackgroundColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*componentColorPalette) BackgroundStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*componentColorPalette) CanvasColor() drawing.Color {
	return drawing.Color{R: 191, G: 254, B: 247, A: 128}
}
func (*componentColorPalette) CanvasStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*componentColorPalette) AxisStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*componentColorPalette) TextColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*componentColorPalette) GetSeriesColor(index int) drawing.Color {
	return drawing.Color{R: 3, G: 247, B: 198, A: 255}
}

func (c *Component) CreateRenderer() fyne.WidgetRenderer { return c }
func (c *Component) Layout(s fyne.Size) {
	pos := fyne.NewPos(0, 0)
	for i, e := range c.entries {
		c.labels[i].Resize(fyne.NewSize(140, c.labels[i].MinSize().Height))
		c.labels[i].Move(pos)
		pos.Y += c.labels[i].MinSize().Height
		e.Resize(fyne.NewSize(140, e.MinSize().Height))
		e.Move(pos)
		pos.Y += e.MinSize().Height
	}
	c.Image.Resize(fyne.NewSize(140, 60))
	c.Image.Move(pos)
}
func (c *Component) MinSize() fyne.Size {
	res := fyne.NewSize(140, 60)
	for _, e := range c.entries {
		res.Height += e.MinSize().Height
	}
	for _, l := range c.labels {
		res.Height += l.MinSize().Height
	}
	return res
}
func (c *Component) Refresh() {
	c.renderChart()
	for _, e := range c.entries {
		e.Refresh()
	}
	for _, l := range c.labels {
		l.Refresh()
	}
}
func (c *Component) Destroy() {}
func (c *Component) Objects() []fyne.CanvasObject {
	res := []fyne.CanvasObject{}
	for i, e := range c.entries {
		res = append(res, c.labels[i])
		res = append(res, e)
	}
	res = append(res, c.Image)
	return res
}
