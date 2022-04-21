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

type component struct {
	widget.BaseWidget
	modeler
	pos             fyne.Position
	currentOverTime []float64
	nodes           [2]int
	currentRange    chart.Range
	chart           *canvas.Image
	labels          []*widget.Label
	entries         []*widget.Entry
}

func newComponent(settings io.Reader, r chart.Range, update func()) *component {
	var c component
	var t string
	var a int
	var b int
	_, err := fmt.Fscanf(settings, "%f %f %s %d %d\n",
		&c.pos.X, &c.pos.Y, &t, &a, &b,
	)
	if err != nil {
		return nil
	}
	if a <= 0 || b <= 0 {
		log.Fatal("unconnected components in the circuit")
	}
	c.nodes[0] = a - 1
	c.nodes[1] = b - 1
	c.setupModeler(t, update)
	c.currentRange = r
	c.currentOverTime = make([]float64, iterations)
	for i := range c.currentOverTime {
		c.currentOverTime[i] = 0
	}
	c.renderChart()
	return &c
}

func (c *component) setupModeler(name string, update func()) {
	c.modeler = newModeler(name)
	c.entries = make([]*widget.Entry, 0)
	c.labels = make([]*widget.Label, 0)
	for k, v := range c.parameters() {
		e := widget.NewEntry()
		e.TextStyle.Monospace = true
		e.SetPlaceHolder(k)
		e.SetText(fmt.Sprintf("%f", v))
		e.OnSubmitted = func(s string) {
			var v float64
			_, err := fmt.Sscanf(s+"\n", "%f\n", &v)
			if err != nil {
				e.SetText(fmt.Sprintf("%f", c.parameters()[k]))
			} else {
				c.updateParameter(k, v)
				update()
			}
		}
		c.entries = append(c.entries, e)
		l := widget.NewLabel(k)
		l.TextStyle.Monospace = true
		c.labels = append(c.labels, l)
	}
}

func (c *component) renderChart() {
	graph := chart.Chart{
		Width:        chartWidth,
		Height:       chartHeight,
		ColorPalette: &componentColorPalette{},
		XAxis:        chart.HideXAxis(),
		YAxis: chart.YAxis{
			Style: chart.Hidden(),
			Range: c.currentRange,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: chart.LinearRange(0, float64(iterations)-1),
				YValues: c.currentOverTime,
			},
		},
	}
	writer := &chart.ImageWriter{}
	graph.Render(chart.PNG, writer)
	img, _ := writer.Image()
	c.chart = canvas.NewImageFromImage(img)
}

type componentColorPalette struct{}

func (*componentColorPalette) BackgroundColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*componentColorPalette) BackgroundStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*componentColorPalette) CanvasColor() drawing.Color {
	return drawing.Color{
		R: currentCanvasR,
		G: currentCanvasG,
		B: currentCanvasB,
		A: currentCanvasA,
	}
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
	return drawing.Color{R: currentR, G: currentG, B: currentB, A: currentA}
}

func (c *component) CreateRenderer() fyne.WidgetRenderer { return c }
func (c *component) Layout(s fyne.Size) {
	pos := fyne.NewPos(0, 0)
	for i, e := range c.entries {
		c.labels[i].Resize(fyne.NewSize(chartWidth, c.labels[i].MinSize().Height))
		c.labels[i].Move(pos)
		pos.Y += c.labels[i].MinSize().Height
		e.Resize(fyne.NewSize(chartWidth, e.MinSize().Height))
		e.Move(pos)
		pos.Y += e.MinSize().Height
	}
	c.chart.Resize(fyne.NewSize(chartWidth, chartHeight))
	c.chart.Move(pos)
}
func (c *component) MinSize() fyne.Size {
	res := fyne.NewSize(chartWidth, chartHeight)
	for _, e := range c.entries {
		res.Height += e.MinSize().Height
	}
	for _, l := range c.labels {
		res.Height += l.MinSize().Height
	}
	return res
}
func (c *component) Refresh() {
	c.renderChart()
	for _, e := range c.entries {
		e.Refresh()
	}
	for _, l := range c.labels {
		l.Refresh()
	}
}
func (c *component) Destroy() {}
func (c *component) Objects() []fyne.CanvasObject {
	res := []fyne.CanvasObject{}
	for i, e := range c.entries {
		res = append(res, c.labels[i])
		res = append(res, e)
	}
	res = append(res, c.chart)
	return res
}
