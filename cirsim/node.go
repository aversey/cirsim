package cirsim

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type Node struct {
	widget.BaseWidget
	Pos             fyne.Position
	VoltageOverTime []float64
	VoltageRange    chart.Range
	AComponents     []*Component
	BComponents     []*Component
	Image           *canvas.Image
}

func NewNode(settings io.Reader, r chart.Range) *Node {
	var n Node
	_, err := fmt.Fscanf(settings, "%f %f\n", &n.Pos.X, &n.Pos.Y)
	if err != nil {
		return nil
	}
	n.VoltageRange = r
	n.VoltageOverTime = make([]float64, 1000)
	for i := range n.VoltageOverTime {
		n.VoltageOverTime[i] = 0
	}
	n.renderChart()
	return &n
}

func (n *Node) renderChart() {
	graph := chart.Chart{
		Width:        140,
		Height:       60,
		ColorPalette: &nodeColorPalette{},
		XAxis:        chart.HideXAxis(),
		YAxis: chart.YAxis{
			Style: chart.Hidden(),
			Range: n.VoltageRange,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: chart.LinearRange(0, 999),
				YValues: n.VoltageOverTime,
			},
		},
	}
	writer := &chart.ImageWriter{}
	graph.Render(chart.PNG, writer)
	img, _ := writer.Image()
	n.Image = canvas.NewImageFromImage(img)
}

type nodeColorPalette struct{}

func (*nodeColorPalette) BackgroundColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*nodeColorPalette) BackgroundStrokeColor() drawing.Color {
	return drawing.ColorTransparent
}
func (*nodeColorPalette) CanvasColor() drawing.Color {
	return drawing.Color{R: 249, G: 254, B: 172, A: 128}
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
	return drawing.Color{R: 247, G: 239, B: 3, A: 255}
}

func (n *Node) CreateRenderer() fyne.WidgetRenderer { return n }
func (n *Node) Layout(s fyne.Size) {
	n.Image.Resize(s)
	n.Image.Move(fyne.NewPos(0, 0))
}
func (n *Node) MinSize() fyne.Size { return n.Image.MinSize() }
func (n *Node) Refresh()           { n.renderChart() }
func (n *Node) Destroy()           {}
func (n *Node) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{n.Image}
}
