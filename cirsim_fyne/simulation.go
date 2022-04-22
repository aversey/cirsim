package cirsim_fyne

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"git.veresov.xyz/aversey/cirsim/cirsim"
	"github.com/wcharczuk/go-chart/v2"
)

const (
	chartWidth     int   = 140
	chartHeight          = 60
	currentCanvasR uint8 = 191
	currentCanvasG       = 254
	currentCanvasB       = 247
	currentCanvasA       = 128
	currentR             = 3
	currentG             = 247
	currentB             = 198
	currentA             = 255
	voltageCanvasR       = 249
	voltageCanvasG       = 254
	voltageCanvasB       = 172
	voltageCanvasA       = 128
	voltageR             = 247
	voltageG             = 239
	voltageB             = 3
	voltageA             = 255
)

type simulation struct {
	sim          cirsim.Simulator
	size         fyne.Size
	nodes        []*node
	components   []*component
	voltageRange chart.ContinuousRange
	currentRange chart.ContinuousRange
	periodEntry  *widget.Entry
	voltageLabel *canvas.Text
	currentLabel *canvas.Text
}

func New() fyne.CanvasObject {
	background := canvas.NewRectangle(color.White)
	circuit := canvas.NewImageFromFile("circuit.svg")
	settings, err := os.Open("circuit")
	if err != nil {
		log.Fatal(err)
	}
	var sim simulation
	sim.voltageRange = chart.ContinuousRange{Min: 0, Max: 0}
	sim.currentRange = chart.ContinuousRange{Min: 0, Max: 0}
	fmt.Fscanf(settings, "%f %f\n\n", &sim.size.Width, &sim.size.Height)
	cont := container.New(&sim, sim.newPanel(settings), background, circuit)
	sim.addNodes(cont, settings)
	sim.addComponents(cont, settings)
	settings.Close()
	components := make([]cirsim.ComponentSettings, len(sim.components))
	for i := range components {
		components[i] = sim.components[i]
	}
	sim.sim = cirsim.New(len(sim.nodes), components)
	sim.periodEntry.SetPlaceHolder(
		fmt.Sprintf("default: %fs", sim.sim.Period()))
	sim.setupComponentModelers()
	sim.update()
	return cont
}

func (sim *simulation) newPanel(settings io.Reader) *fyne.Container {
	sim.voltageLabel = canvas.NewText(
		fmt.Sprintf(" %e < voltage < %e ",
			sim.voltageRange.Min, sim.voltageRange.Max),
		color.RGBA{R: voltageR, G: voltageG, B: voltageB, A: voltageA},
	)
	sim.voltageLabel.TextStyle.Monospace = true
	sim.currentLabel = canvas.NewText(
		fmt.Sprintf(" %e < current < %e ",
			sim.currentRange.Min, sim.currentRange.Max),
		color.RGBA{R: currentR, G: currentG, B: currentB, A: currentA},
	)
	sim.currentLabel.TextStyle.Monospace = true
	periodLabel := widget.NewLabel("Period")
	periodLabel.TextStyle.Monospace = true
	sim.periodEntry = widget.NewEntry()
	sim.periodEntry.TextStyle.Monospace = true
	sim.periodEntry.OnSubmitted = sim.updatePeriod
	return container.NewHBox(
		sim.voltageLabel,
		sim.currentLabel,
		layout.NewSpacer(),
		periodLabel,
		container.New(&entryLayout{}, sim.periodEntry),
	)
}

func (sim *simulation) addNodes(cont *fyne.Container, settings io.Reader) {
	n := newNode(settings, &sim.voltageRange)
	for n != nil {
		sim.nodes = append(sim.nodes, n)
		cont.Add(n)
		n = newNode(settings, &sim.voltageRange)
	}
}

func (sim *simulation) addComponents(cont *fyne.Container, settings io.Reader) {
	c := newComponent(settings, &sim.currentRange)
	for c != nil {
		sim.components = append(sim.components, c)
		cont.Add(c)
		c = newComponent(settings, &sim.currentRange)
	}
}

func (sim *simulation) setupComponentModelers() {
	for i := range sim.components {
		sim.components[i].setupModeler(
			sim.sim.ModelerOfComponent(i), sim.update)
	}
}

func (sim *simulation) update() {
	sim.sim.Simulate()
	sim.voltageRange.Min, sim.voltageRange.Max = sim.sim.VoltageRange()
	sim.currentRange.Min, sim.currentRange.Max = sim.sim.CurrentRange()
	sim.voltageLabel.Text = fmt.Sprintf(" %e < voltage < %e ",
		sim.voltageRange.Min, sim.voltageRange.Max)
	sim.currentLabel.Text = fmt.Sprintf(" %e < current < %e ",
		sim.currentRange.Min, sim.currentRange.Max)
	sim.voltageLabel.Refresh()
	sim.currentLabel.Refresh()
	for i, n := range sim.nodes {
		n.renderChart(sim.sim.VoltagesOfNode(i))
		n.Refresh()
	}
	for i, c := range sim.components {
		c.renderChart(sim.sim.CurrentsOfComponent(i))
		c.Refresh()
	}
}

func (sim *simulation) updatePeriod(period string) {
	var periodVal float64
	_, err := fmt.Sscanf(period+"\n", "%f\n", periodVal)
	if err != nil {
		sim.periodEntry.SetText(fmt.Sprintf("%f", sim.sim.Period()))
	} else {
		sim.sim.SetPeriod(periodVal)
		sim.update()
	}
}

func (l *simulation) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return objects[2].MinSize()
}
func (l *simulation) Layout(obs []fyne.CanvasObject, size fyne.Size) {
	var p fyne.Position
	var s fyne.Size
	panelHeight := obs[0].MinSize().Height
	H := size.Height - panelHeight
	W := size.Width
	h := l.size.Height
	w := l.size.Width
	var scale float32
	if H/W > h/w {
		p = fyne.NewPos(0, (H-W*h/w)/2+panelHeight)
		s = fyne.NewSize(W, W*h/w)
		scale = W / w
	} else {
		p = fyne.NewPos((W-H*w/h)/2, panelHeight)
		s = fyne.NewSize(H*w/h, H)
		scale = H / h
	}
	obs[0].Resize(fyne.NewSize(W, panelHeight))
	obs[0].Move(fyne.NewPos(0, 0))
	obs[1].Resize(size)
	obs[1].Move(fyne.NewPos(0, panelHeight))
	obs[2].Resize(s)
	obs[2].Move(p)
	const chartWidth = float32(chartWidth)
	const chartHeight = float32(chartHeight)
	for i, n := range l.nodes {
		obs[3+i].Resize(fyne.NewSize(chartWidth, chartHeight))
		shift := fyne.NewPos(
			p.X+n.pos.X*scale-chartWidth/2.0,
			p.Y+n.pos.Y*scale-chartHeight/2.0,
		)
		obs[3+i].Move(shift)
	}
	for i, c := range l.components {
		ms := obs[3+len(l.nodes)+i].MinSize()
		obs[3+len(l.nodes)+i].Resize(ms)
		shift := fyne.NewPos(
			p.X+c.pos.X*scale-chartWidth/2.0,
			p.Y+c.pos.Y*scale-ms.Height+chartHeight/2.0,
		)
		obs[3+len(l.nodes)+i].Move(shift)
	}
}

type entryLayout struct{}

func (*entryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(
		objects[0].MinSize().Width*6,
		objects[0].MinSize().Height,
	)
}
func (*entryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}
