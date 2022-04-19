package cirsim

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/wcharczuk/go-chart/v2"
	"gonum.org/v1/gonum/mat"
)

type Simulation struct {
	Period       float64
	Size         fyne.Size
	Nodes        []*Node
	Components   []*Component
	VoltageRange chart.ContinuousRange
	CurrentRange chart.ContinuousRange
	periodEntry  *widget.Entry
	voltageLabel *canvas.Text
	currentLabel *canvas.Text
}

func New() fyne.CanvasObject {
	background := canvas.NewRectangle(color.White)
	image := canvas.NewImageFromFile("circuit.svg")
	file, err := os.Open("circuit")
	if err != nil {
		log.Fatal(err)
	}
	var sim Simulation
	sim.Period = 0.01
	sim.VoltageRange = chart.ContinuousRange{Min: 0, Max: 0}
	sim.CurrentRange = chart.ContinuousRange{Min: 0, Max: 0}
	fmt.Fscanf(file, "%f %f\n\n", &sim.Size.Width, &sim.Size.Height)
	sim.voltageLabel = canvas.NewText(
		fmt.Sprintf(" %e < voltage < %e ",
			sim.VoltageRange.Min, sim.VoltageRange.Max),
		color.RGBA{R: 247, G: 239, B: 3, A: 255},
	)
	sim.voltageLabel.TextStyle.Monospace = true
	sim.currentLabel = canvas.NewText(
		fmt.Sprintf(" %e < current < %e ",
			sim.CurrentRange.Min, sim.CurrentRange.Max),
		color.RGBA{R: 3, G: 247, B: 198, A: 255},
	)
	sim.currentLabel.TextStyle.Monospace = true
	periodLabel := widget.NewLabel("Period")
	periodLabel.TextStyle.Monospace = true
	sim.periodEntry = widget.NewEntry()
	sim.periodEntry.TextStyle.Monospace = true
	sim.periodEntry.SetPlaceHolder("default: 0.01s")
	sim.periodEntry.OnSubmitted = sim.updatePeriod
	c := container.New(&sim,
		container.NewHBox(
			sim.voltageLabel,
			sim.currentLabel,
			layout.NewSpacer(),
			periodLabel,
			container.New(&entryLayout{}, sim.periodEntry),
		),
		background,
		image,
	)
	for n := NewNode(file, &sim.VoltageRange); n != nil; n = NewNode(file, &sim.VoltageRange) {
		sim.Nodes = append(sim.Nodes, n)
		c.Add(n)
	}
	for component := NewComponent(file, &sim.CurrentRange, sim.Nodes, sim.simulate); component != nil; component = NewComponent(file, &sim.CurrentRange, sim.Nodes, sim.simulate) {
		sim.Components = append(sim.Components, component)
		c.Add(component)
	}
	file.Close()
	sim.simulate()
	return c
}

func (sim *Simulation) simulate() {
	for _, n := range sim.Nodes {
		for i := range n.VoltageOverTime {
			n.VoltageOverTime[i] = 0
		}
	}
	for _, c := range sim.Components {
		for i := range c.CurrentOverTime {
			c.CurrentOverTime[i] = 0
		}
	}
	maxv := 0.0
	minv := 0.0
	maxc := 0.0
	minc := 0.0
	for i := 0; i != 1000; i++ {
		N := len(sim.Nodes)
		for refine := 0; refine != 10; refine++ {
			m := mat.NewDense(N+1, N, nil)
			v := mat.NewVecDense(N+1, nil)
			for j, n := range sim.Nodes {
				for _, c := range sim.Components {
					if c.ANode == n {
						v.SetVec(j, v.AtVec(j)-c.Model.ModelCurrent(
							float64(i)*sim.Period/1000, sim.Period/1000,
							c.BNode.VoltageOverTime[i]-n.VoltageOverTime[i],
							c.CurrentOverTime[i]))
						d := 0
						for k, v := range sim.Nodes {
							if c.BNode == v {
								d = k
								break
							}
						}
						m.Set(j, d, m.At(j, d)-
							c.Model.ModelConductance(
								float64(i)*sim.Period/1000, sim.Period/1000,
								c.BNode.VoltageOverTime[i]-n.VoltageOverTime[i],
								c.CurrentOverTime[i]))
						m.Set(j, j, m.At(j, j)+
							c.Model.ModelConductance(
								float64(i)*sim.Period/1000, sim.Period/1000,
								c.BNode.VoltageOverTime[i]-n.VoltageOverTime[i],
								c.CurrentOverTime[i]))
					} else if c.BNode == n {
						v.SetVec(j, v.AtVec(j)+c.Model.ModelCurrent(
							float64(i)*sim.Period/1000, sim.Period/1000,
							n.VoltageOverTime[i]-c.ANode.VoltageOverTime[i],
							c.CurrentOverTime[i]))
						d := 0
						for k, v := range sim.Nodes {
							if c.ANode == v {
								d = k
								break
							}
						}
						m.Set(j, d, m.At(j, d)-
							c.Model.ModelConductance(
								float64(i)*sim.Period/1000, sim.Period/1000,
								n.VoltageOverTime[i]-c.ANode.VoltageOverTime[i],
								c.CurrentOverTime[i]))
						m.Set(j, j, m.At(j, j)+
							c.Model.ModelConductance(
								float64(i)*sim.Period/1000, sim.Period/1000,
								n.VoltageOverTime[i]-c.ANode.VoltageOverTime[i],
								c.CurrentOverTime[i]))
					}
				}
			}
			r := make([]float64, N)
			for j := 0; j != N-1; j++ {
				r[j] = 0
			}
			r[N-1] = 1
			m.SetRow(N, r)
			v.SetVec(N, 0)
			res := mat.NewVecDense(N, nil)
			res.SolveVec(m, v)
			for j, n := range sim.Nodes {
				n.VoltageOverTime[i] = res.AtVec(j)
			}
			for _, c := range sim.Components {
				c.CurrentOverTime[i] = (c.ANode.VoltageOverTime[i]-c.BNode.VoltageOverTime[i])*
					c.Model.ModelConductance(
						float64(i)*sim.Period/1000, sim.Period/1000,
						c.BNode.VoltageOverTime[i]-c.ANode.VoltageOverTime[i],
						c.CurrentOverTime[i]) +
					c.Model.ModelCurrent(
						float64(i)*sim.Period/1000, sim.Period/1000,
						c.BNode.VoltageOverTime[i]-c.ANode.VoltageOverTime[i],
						c.CurrentOverTime[i])
			}
		}
		if i == 999 {
			break
		}
		for _, n := range sim.Nodes {
			n.VoltageOverTime[i+1] = n.VoltageOverTime[i]
			if n.VoltageOverTime[i] > maxv {
				maxv = n.VoltageOverTime[i]
			} else if n.VoltageOverTime[i] < minv {
				minv = n.VoltageOverTime[i]
			}
		}
		for _, c := range sim.Components {
			c.CurrentOverTime[i+1] = c.CurrentOverTime[i]
			if c.CurrentOverTime[i] > maxc {
				maxc = c.CurrentOverTime[i]
			} else if c.CurrentOverTime[i] < minc {
				minc = c.CurrentOverTime[i]
			}
		}
	}
	sim.VoltageRange.Max = maxv
	sim.VoltageRange.Min = minv
	sim.CurrentRange.Max = maxc
	sim.CurrentRange.Min = minc
	sim.voltageLabel.Text = fmt.Sprintf(" %e < voltage < %e ",
		sim.VoltageRange.Min, sim.VoltageRange.Max)
	sim.currentLabel.Text = fmt.Sprintf(" %e < current < %e ",
		sim.CurrentRange.Min, sim.CurrentRange.Max)
	sim.voltageLabel.Refresh()
	sim.currentLabel.Refresh()
	for _, n := range sim.Nodes {
		n.Refresh()
	}
	for _, c := range sim.Components {
		c.Refresh()
	}
}

func (sim *Simulation) updatePeriod(period string) {
	_, err := fmt.Sscanf(period+"\n", "%f\n", &sim.Period)
	if err != nil {
		sim.periodEntry.SetText(fmt.Sprintf("%f", sim.Period))
	} else {
		sim.simulate()
	}
}

func (l *Simulation) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := fyne.NewSize(0, 0)
	for _, i := range objects {
		switch o := i.(type) {
		case *canvas.Image:
			s = o.MinSize()
		default:
		}
	}
	return s
}

func (l *Simulation) Layout(obs []fyne.CanvasObject, size fyne.Size) {
	var p fyne.Position
	var s fyne.Size
	panel_height := obs[0].MinSize().Height
	H := size.Height - panel_height
	W := size.Width
	h := l.Size.Height
	w := l.Size.Width
	var scale float32
	if H/W > h/w {
		p = fyne.NewPos(0, (H-W*h/w)/2+panel_height)
		s = fyne.NewSize(W, W*h/w)
		scale = W / w
	} else {
		p = fyne.NewPos((W-H*w/h)/2, panel_height)
		s = fyne.NewSize(H*w/h, H)
		scale = H / h
	}
	obs[0].Resize(fyne.NewSize(W, panel_height))
	obs[0].Move(fyne.NewPos(0, 0))
	obs[1].Resize(size)
	obs[1].Move(fyne.NewPos(0, panel_height))
	obs[2].Resize(s)
	obs[2].Move(p)
	for i, n := range l.Nodes {
		obs[3+i].Resize(fyne.NewSize(140, 60))
		shift := fyne.NewPos(p.X+n.Pos.X*scale-70, p.Y+n.Pos.Y*scale-30)
		obs[3+i].Move(shift)
	}
	for i, c := range l.Components {
		ms := obs[3+len(l.Nodes)+i].MinSize()
		obs[3+len(l.Nodes)+i].Resize(ms)
		shift := fyne.NewPos(p.X+c.Pos.X*scale-70, p.Y+c.Pos.Y*scale-ms.Height+30)
		obs[3+len(l.Nodes)+i].Move(shift)
	}
}

type entryLayout struct{}

func (*entryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(
		objects[0].MinSize().Width*4,
		objects[0].MinSize().Height,
	)
}
func (*entryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}
