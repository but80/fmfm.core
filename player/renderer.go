package player

import (
	"fmt"
	"sync"

	"github.com/gordonklaus/portaudio"
	"github.com/xlab/closer"
)

type Renderer struct {
	Parameters portaudio.StreamParameters
	stream     *portaudio.Stream
}

var portautioInitOnce = sync.Once{}

func NewRenderer() *Renderer {
	renderer := &Renderer{}
	portautioInitOnce.Do(func() {
		portaudio.Initialize()
		closer.Bind(func() {
			portaudio.Terminate()
		})
	})

	h, err := portaudio.DefaultHostApi()
	if err != nil {
		panic(err)
	}
	selectedDevinfo := h.DefaultOutputDevice

	// var selectedDevinfo *portaudio.DeviceInfo
	// devinfos, err := portaudio.Devices()
	// if err != nil {
	// 	panic(err)
	// }
	// for _, devinfo := range devinfos {
	// 	if 0 < devinfo.MaxOutputChannels {
	// 		if deviceName == devinfo.Name {
	// 			selectedDevinfo = devinfo
	// 		}
	// 	}
	// }
	// if selectedDevinfo == nil {
	// 	panic("device not found")
	// }

	renderer.Parameters = portaudio.HighLatencyParameters(nil, selectedDevinfo)
	// params := portaudio.StreamParameters{
	// 	Output: portaudio.StreamDeviceParameters{
	// 		Device: selectedDevinfo,
	// 		Channels: selectedDevinfo.MaxOutputChannels,
	// 		Latency: selectedDevinfo.DefaultHighOutputLatency,
	// 	},
	// 	SampleRate: sampleRate,
	// 	FramesPerBuffer: 0,
	// }

	return renderer
}

func (renderer *Renderer) Start(processor func() (float64, float64)) {
	var err error
	renderer.stream, err = portaudio.OpenStream(renderer.Parameters, func(out [][]float32) {
		for i := range out[0] {
			l, r := processor()
			out[0][i] = float32(l)
			out[1][i] = float32(r)
		}
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sample rate: %f\n", renderer.stream.Info().SampleRate)
	fmt.Printf("Output latency: %s\n", renderer.stream.Info().OutputLatency.String())

	err = renderer.stream.Start()
	if err != nil {
		panic(err)
	}
}
