package backlight

import (
	"fmt"

	"barista.run/bar"
	"barista.run/base/value"
	"barista.run/outputs"

	"github.com/fsnotify/fsnotify"
)

// Module is the barista module for controlling the screen backlight of laptops.
type Module struct {
	formatFunc value.Value
	kernel     string
}

// New creates a new instance of Module.
func New(kernel string) *Module {
	m := new(Module)
	m.formatFunc.Set(func(b *Backlight) bar.Output {
		return outputs.Textf("%d%%", b.Percent())
	})
	m.kernel = kernel

	return m
}

// Stream is the barista stream function to update the status bar state.
func (m *Module) Stream(s bar.Sink) {
	// Get the screen brightness at the beginning to set the initial state
	b := NewBacklight(m.kernel)
	err := m.updateOutput(s, b)
	if err != nil {
		s.Error(err)
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.Error(err)
		return
	}
	defer watcher.Close()

	// Watch the actual_brightness file to monitor the current brightness
	err = watcher.Add(fmt.Sprintf("/sys/class/backlight/%s/actual_brightness", m.kernel))
	if err != nil {
		s.Error(err)
		return
	}

	for {
		select {
		// Listen to filesystem events
		case event, ok := <-watcher.Events:
			if ok && event.Op&fsnotify.Write == fsnotify.Write {
				// Update the backlight info if the brightness changed
				err := m.updateOutput(s, b)
				if err != nil {
					s.Error(err)
					return
				}
			}
		// Listen to errors of the filesystem watcher
		case err := <-watcher.Errors:
			s.Error(err)
			return
		}
	}

}

func (m *Module) updateOutput(s bar.Sink, b *Backlight) error {
	err := b.Get()
	if err != nil {
		return err
	}

	format := m.formatFunc.Get().(func(b *Backlight) bar.Output)
	s.Output(outputs.Group(format(b)).OnClick(clickHandler(b)))

	return nil
}

// Output sets the format function to enable custom output formats.
func (m *Module) Output(format func(b *Backlight) bar.Output) *Module {
	m.formatFunc.Set(format)
	return m
}

func clickHandler(b *Backlight) func(bar.Event) {
	return func(e bar.Event) {
		step := b.Max / 100
		if step == 0 {
			step = 1
		}
		if e.Button == bar.ScrollUp {
			b.SetBrightness(b.Bri + step)
		}
		if e.Button == bar.ScrollDown {
			b.SetBrightness(b.Bri - step)
		}
	}
}
