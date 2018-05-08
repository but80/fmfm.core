# fmFM

[![Build Status](https://travis-ci.org/but80/fmfm.svg?branch=master)](https://travis-ci.org/but80/fmfm)
[![Go Report Card](https://goreportcard.com/badge/github.com/but80/fmfm)](https://goreportcard.com/report/github.com/but80/fmfm)
[![Godoc](https://godoc.org/github.com/but80/fmfm?status.svg)](http://godoc.org/github.com/but80/fmfm)

**WORK IN PROGRESS**

**fmFM** (Fake Mobile FM synth) is a YAMAHA MA-5 (YMU765) / YMF825 clone software FM synthesizer.

Most of this code is based on [doomjs/opl3](https://github.com/doomjs/opl3).

# Requirements

- macOS
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

# Usage

```
go run cmd/fmfm-cli/main.go
```

- A voice library (`.vm5`) must be placed on `voice/default.vm5` before running.
- The IAC virtual MIDI port named `IAC YAMAHA Virtual MIDI Device 0` must be created before running.
  fmFM receives the MIDI messages via this port.
