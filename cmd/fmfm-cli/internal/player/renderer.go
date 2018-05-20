package player

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/xlab/closer"
)

// Renderer は、波形をレンダリングしてオーディオデバイスに出力します。
// TODO: rename
type Renderer struct {
	Parameters portaudio.StreamParameters
	stream     *portaudio.Stream
}

var portautioInitOnce = sync.Once{}

// NewRenderer は、新しいRendererを作成します。
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
	fmt.Printf("Device: %s\n", selectedDevinfo.Name)

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
	// 		channels: selectedDevinfo.MaxOutputChannels,
	// 		Latency: selectedDevinfo.DefaultHighOutputLatency,
	// 	},
	// 	SampleRate: sampleRate,
	// 	FramesPerBuffer: 0,
	// }

	return renderer
}

// Start は、processor によって生成される波形のオーディオデバイスへの出力を開始します。
func (renderer *Renderer) Start(processor func() (float64, float64), controller func(int)) {
	startTime := time.Now()
	maxLevel := 32766.0 / 32767.0

	var err error
	renderer.stream, err = portaudio.OpenStream(renderer.Parameters, func(out [][]float32) {
		// midiLatency := float64(renderer.stream.Info().OutputLatency) / float64(time.Millisecond)
		sampleLen := 1000.0 / renderer.Parameters.SampleRate
		midiLatency := float64(len(out[0])) * sampleLen
		now := float64(time.Since(startTime)) / float64(time.Millisecond)
		for i := range out[0] {
			now += sampleLen
			controller(int(now - midiLatency))

			l, r := processor()
			out[0][i] = float32(l)
			out[1][i] = float32(r)
			if maxLevel < l || maxLevel < r {
				if maxLevel < l {
					maxLevel = l
				}
				if maxLevel < r {
					maxLevel = r
				}
				db := math.Log10(maxLevel) * 20.0
				fmt.Printf("Clipping occurred: %2.1f\n", db)
			}
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
