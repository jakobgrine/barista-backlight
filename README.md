# barista-backlight
A module for the [barista](https://barista.run) status bar to display and control the screen brightness.

## Usage
~~~go
package main

import (
    "barista.run"
    "barista.run/bar"
    "barista.run/outputs"
  
    "github.com/jakobgrine/barista-backlight"
)

func main() {
    barista.Add(backlight.New("acpi_video0").Output(func(b *backlight.Backlight) bar.Output {
        return outputs.Textf("%d%%", b.Percent())
    }))
  
    panic(barista.Run())
}
~~~
