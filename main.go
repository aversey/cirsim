package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"git.veresov.xyz/aversey/cirsim/cirsim_fyne"
)

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Circuit Simulator")
	w.Resize(fyne.NewSize(1280, 720))
	w.SetIcon(theme.SettingsIcon())
	w.SetContent(cirsim_fyne.New())
	w.ShowAndRun()
}
