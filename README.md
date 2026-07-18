# Simple Deflicker
A minimalist, easy to use tool for deflickering image sequences such as timelapses.

![Short Demo](demo_church.gif)

## What is timelapse flickering?
Timelapse flickering can occur if one or more settings of the camera have been left on "auto" which causes it to randomly switch between two settings (for example shutter speeds).

## How to use this software
* Download the latest version from the [releases page](https://github.com/StruffelProductions/simple-deflicker/releases). Prebuilt binaries are provided for Windows, and macOS builds are CLI-only.
* Execute `simple-deflicker.exe` on Windows to use the GUI. Check the console for error messages.
![image](https://user-images.githubusercontent.com/31403260/115123359-f2bbe400-9fbc-11eb-84d7-29615c5030fb.png)

## CLI usage
Simple Deflicker can run without the GUI by passing a source and destination directory:

```bash
simple-deflicker -source "/path/to/input" -destination "/path/to/output"
```

Optional flags:

```bash
simple-deflicker -source "/path/to/input" -destination "/path/to/output" -rollingAverage 15 -jpegCompression 95 -threads 8
```

Build a CLI-only binary:

```bash
go build -tags cli -o simple-deflicker
```

macOS builds are CLI-only:

```bash
GOOS=darwin GOARCH=arm64 go build -o simple-deflicker-macos-arm64
GOOS=darwin GOARCH=amd64 go build -o simple-deflicker-macos-amd64
```

On platforms built with the `cli` tag, the GUI is disabled and `-source` plus `-destination` are required.


## Current limitations of the tool
* Only JPG and PNG (8bit) are supported
* JPGs will always be saved with a compression setting of 95
* All metadata present in the source files will not be copied over.
* The software can only fix global flicker. It can not deal with rolling flicker (caused by certain indoor lighting conditions).

## How does the deflickering work?
The current implementation uses a technique called [histogram matching](https://en.wikipedia.org/wiki/Histogram_matching). It basically creates a list of how often a certain brighness (or rather every individual brightness level) appears, creates a [rolling average](https://en.wikipedia.org/wiki/Moving_average) to allow for gradual brightness changes (for example in a day to night transition) and finally shifts the brightness to match the "smoothed out" brightness levels.

## How is the software structured? (only important for developers, not for users)
The software uses several other packages:
* [Imaging](https://github.com/disintegration/imaging) for loading, saving and manipulating image files.
* [dialog](https://github.com/sqweek/dialog) for creating the dialog boxes and file selection windows.
* [uiprogress](https://github.com/gosuri/uiprogress) for creating the progress bars in the console.
* [nucular](https://github.com/aarzilli/nucular) for the GUI.
