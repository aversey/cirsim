package cirsim_fyne

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type node struct {
	widget.BaseWidget
	pos          fyne.Position
	voltageRange chart.Range
	chart        *canvas.Image
}

func newNode(settings io.Reader, r chart.Range) *node {
	var n node
	_, err := fmt.Fscanf(settings, "%f %f\n", &n.pos.X, &n.pos.Y)
	if err != nil {
		return nil
	}
	n.voltageRange = r
	return &n
}

func (n *node) renderChart(voltageOverTime []float64) {
	graph := chart.Chart{
		Width:        chartWidth,
		Height:       chartHeight,
		ColorPalette: &nodeColorPalette{},
		XAxis:        chart.HideXAxis(),
		YAxis: chart.YAxis{
			Style: chart.Hidden(),
			Range: n.voltageRange,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: chart.LinearRange(0, float64(len(voltageOverTime))-1),
				YValues: voltageOverTime,
			},
		},
	}
	writer := &chart.ImageWriter{}
	graph.Render(chart.PNG, writer)
	img, _ := writer.Image()
	n.chart = canvas.NewImageFromImage(img)
}

type nodeColorPalette struct{}

func (*nodeColorPalette) BackgroundColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*nodeColorPalette) BackgroundStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*nodeColorPalette) CanvasColor() drawing.Color {
	return drawing.Color{
		R: voltageCanvasR,
		G: voltageCanvasG,
		B: voltageCanvasB,
		A: voltageCanvasA,
	}
}
func (*nodeColorPalette) CanvasStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*nodeColorPalette) AxisStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*nodeColorPalette) TextColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*nodeColorPalette) GetSeriesColor(index int) drawing.Color {
	return drawing.Color{R: voltageR, G: voltageG, B: voltageB, A: voltageA}
}

func (n *node) CreateRenderer() fyne.WidgetRenderer { return n }
func (n *node) Layout(s fyne.Size) {
	n.chart.Resize(s)
	n.chart.Move(fyne.NewPos(0, 0))
}
func (n *node) MinSize() fyne.Size { return n.chart.MinSize() }
func (n *node) Refresh()           {}
func (n *node) Destroy()           {}
func (n *node) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{n.chart}
}
