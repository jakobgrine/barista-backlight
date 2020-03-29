package backlight

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"barista.run/bar"
	"barista.run/base/value"
	"barista.run/outputs"

	"github.com/fsnotify/fsnotify"
)

// BacklightInfo holds information about the current state of the screen backlight
type BacklightInfo struct {
	bri int
	max int
}

// Percent returns the brightness in percent of the maximum value
func (b *BacklightInfo) Percent() int {
	if b.max == 0 {
		return 0
	}
	return int(float64(b.bri) / float64(b.max) * 100.0)
}

// Module is the brightness module for the barista status bar
type Module struct {
	formatFunction value.Value
	kernel         string
}

// New creates a new instance of Module
func New(kernel string) *Module {
	m := new(Module)
	m.formatFunction.Set(func(b *BacklightInfo) bar.Output {
		return outputs.Text(fmt.Sprintf("%d%%", b.Percent()))
	})
	m.kernel = kernel

	return m
}

// Stream is the barista stream function to update the status bar state
func (m *Module) Stream(s bar.Sink) {
	// Get the screen brightness at the beginning to set the initial state
	info, err := m.getBacklightInfo()
	if err != nil {
		s.Error(err)
		return
	}
	format := m.formatFunction.Get().(func(b *BacklightInfo) bar.Output)
	s.Output(format(info))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.Error(err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					s.Error(errors.New("fsnotify.Watcher not ok"))
					done <- true
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// Update the backlight info if the brightness changed
					info, err := m.getBacklightInfo()
					if err != nil {
						s.Error(err)
						done <- true
					}

					format := m.formatFunction.Get().(func(b *BacklightInfo) bar.Output)
					s.Output(format(info))
				}
			case err := <-watcher.Errors:
				s.Error(err)
				done <- true
			}
		}
	}()

	// Add the actual_brightness file to the filesystem watcher to monitor the current brightness
	err = watcher.Add("/sys/class/backlight/intel_backlight/actual_brightness")
	if err != nil {
		s.Error(err)
		return
	}

	<-done
}

// Output sets the format function to enable custom output formats
func (m *Module) Output(format func(b *BacklightInfo) bar.Output) *Module {
	m.formatFunction.Set(format)
	return m
}

func readIntFromFile(file string) (int, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}

	data := strings.Replace(string(bytes), "\n", "", 1)

	n, err := strconv.Atoi(data)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (m *Module) getBacklightInfo() (*BacklightInfo, error) {
	max, err := readIntFromFile("/sys/class/backlight/intel_backlight/max_brightness")
	if err != nil {
		return nil, err
	}

	bri, err := readIntFromFile("/sys/class/backlight/intel_backlight/actual_brightness")
	if err != nil {
		return nil, err
	}

	return &BacklightInfo{bri, max}, nil
}
