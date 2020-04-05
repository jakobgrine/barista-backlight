package backlight

import (
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

// Backlight holds information about the current state of the screen backlight.
type Backlight struct {
	Bri    int
	Max    int
	Kernel string
}

// NewBacklight creates a new instance of Backlight.
func NewBacklight(kernel string) *Backlight {
	return &Backlight{0, 0, kernel}
}

// SetBrightness sets the screen brightness.
func (b *Backlight) SetBrightness(value int) error {
	file := fmt.Sprintf("/sys/class/backlight/%s/brightness", b.Kernel)
	bytes := []byte(strconv.Itoa(value))
	return ioutil.WriteFile(file, bytes, 0644)
}

// Get updates the Bri and Max values after reading the respective files.
func (b *Backlight) Get() error {
	max, err := b.readValue("max_brightness")
	if err != nil {
		return err
	}
	b.Max = max

	bri, err := b.readValue("actual_brightness")
	if err != nil {
		return err
	}
	b.Bri = bri

	return nil
}

func (b *Backlight) readValue(name string) (int, error) {
	file := fmt.Sprintf("/sys/class/backlight/%s/%s", b.Kernel, name)
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}

	dat := strings.Replace(string(bytes), "\n", "", 1)

	n, err := strconv.Atoi(dat)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Fraction returns the brightness as a fraction of the maximum value.
func (b *Backlight) Fraction() float64 {
	if b.Max == 0 {
		return 0
	}
	return float64(b.Bri) / float64(b.Max)
}

// Percent returns the brightness in percent of the maximum value.
func (b *Backlight) Percent() int {
	return int(math.Round(b.Fraction() * 100.0))
}
