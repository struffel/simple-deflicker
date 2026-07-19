# Simple Deflicker
A minimalist, easy to use tool for deflickering image sequences such as timelapses.

![Short Demo](demo_church.gif)

## What is timelapse flickering?
Timelapse flickering can occur if one or more settings of the camera have been left on "auto" which causes it to randomly switch between two settings (for example shutter speeds).

## How to use this software
* Download the latest version from the [releases page](https://github.com/struffel/simple-deflicker/releases). Prebuilt binaries are provided only for Windows at this time.
* Execute `simple-deflicker.exe` to use the GUI or `simple-deflicker-cli.exe` to use the CLI version.

## CLI usage
Simple Deflicker can run without the GUI by passing a source and destination directory:

```bash
simple-deflicker-cli -source "/path/to/input" -destination "/path/to/output"
```

Optional flags:

```bash
simple-deflicker-cli -source "/path/to/input" -destination "/path/to/output" -rollingAverage 15 -format png -jpegQuality 95
```

## Building from source
The GUI and CLI are separate binaries, built from `./cmd/gui` and `./cmd/cli` respectively:

```bash
go build -o simple-deflicker ./cmd/gui
go build -o simple-deflicker-cli ./cmd/cli
```

## Current limitations of the tool
* Only JPG and PNG (8bit) are supported
* JPGs will always be saved with a compression setting of 95
* All metadata present in the source files will not be copied over.
* The software can only fix global flicker. It can not deal with rolling flicker (caused by certain indoor lighting conditions).

## How does the deflickering work?
The current implementation uses a technique called [histogram matching](https://en.wikipedia.org/wiki/Histogram_matching). It basically creates a list of how often a certain brighness (or rather every individual brightness level) appears, creates a [rolling average](https://en.wikipedia.org/wiki/Moving_average) to allow for gradual brightness changes (for example in a day to night transition) and finally shifts the brightness to match the "smoothed out" brightness levels.
