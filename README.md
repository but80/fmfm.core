# fmFM

[![Go Report Card](https://goreportcard.com/badge/github.com/but80/fmfm)](https://goreportcard.com/report/github.com/but80/fmfm)
[![Godoc](https://godoc.org/github.com/but80/fmfm?status.svg)](http://godoc.org/github.com/but80/fmfm)

**WORK IN PROGRESS**

YAMAHA MA-5 / YMF825 clone software FM synthesizer

- Most of this code is based on [doomjs/opl3](https://github.com/doomjs/opl3).

## Requirements

- [PortMIDI](http://portmedia.sourceforge.net/portmidi/)
  
  ```
  # On macOS
  brew install portmidi
  ```
- [PortAudio](http://www.portaudio.com/)
  
  ```
  # On macOS
  brew install portaudio
  ```

## Usage

```
go run main.go
```

- A voice library (`.vm5`) must be placed on `voice/default.vm5` before running.
- Receives MIDI messages via the port named `IAC YAMAHA Virtual MIDI Device 0`.
