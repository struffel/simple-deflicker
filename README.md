# Simple Deflicker
A minimalist, lightning-fast and easy to use tool for deflickering image sequences such as timelapses.
It's still in its early stages of development.

![Short Demo](demo_church.gif)

## What is timelapse flickering?
Timelapse flickering can occur if one or more settings of the camera have been left on "auto" which causes it to randomly switch between two settings (for example shutter speeds).

## How to use this software
* Download the latest version from the [releases page](https://github.com/StruffelProductions/simple-deflicker/releases). The compiled binary is only available for windows at this point
* Execute simple-deflicker.exe. Starting with v0.3.0 there will be a (very basic) GUI to enter all the settings. Check the console for error messages.
![image](https://user-images.githubusercontent.com/31403260/115123359-f2bbe400-9fbc-11eb-84d7-29615c5030fb.png)


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

