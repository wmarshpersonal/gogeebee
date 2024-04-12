package main

import (
	"log/slog"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"

	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/gb"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

type Game struct {
	gb   *gb.GB
	sync synchronization
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func initGame(romData []byte) (*Game, error) {
	mbc, err := cartridge.Load(romData)
	if err != nil {
		return nil, err
	}

	g := &Game{
		gb: gb.NewDMG(mbc),
		sync: synchronization{
			cond:         sync.NewCond(&sync.Mutex{}),
			audioSamples: make([]byte, 0, sampleRate),
		},
	}

	const chs = 2
	op := &oto.NewContextOptions{}
	op.SampleRate = sampleRate
	op.ChannelCount = chs
	op.Format = oto.FormatFloat32LE

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, err
	}
	<-readyChan

	str := &audioStream{&g.sync}
	p := otoCtx.NewPlayer(str)
	time.AfterFunc(math.MaxInt64, func() { runtime.KeepAlive(p) })

	const bufferSize = 12288
	p.SetBufferSize(12288)

	p.Play()
	// defer p.Close()

	go func() {
		var (
			apuSamples []uint8
			samples    []byte
			frame      ppu.PixelBuffer
		)
		for {
			g.gb.ProcessJoypad(ReadButtons(), ReadDirections())
			drawn := g.gb.RunFor(gb.TCyclesPerSecond*(bufferSize/8)/sampleRate, &frame, &apuSamples)
			var dropped bool
			if drawn > 0 {
				dropped = g.sync.addFrame(frame)
			}
			samples = samples[:0]
			const samplesPerFrame = bufferSize / 4 / chs
			var inc = float64(len(apuSamples)) / samplesPerFrame
			var i int
			for i = 0; i < samplesPerFrame; i++ {
				sample := (float32(apuSamples[int(float64(i)*inc)])/0xF - 0.5)
				bits := math.Float32bits(sample)
				samples = append(samples, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))
				samples = append(samples, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))
			}
			apuSamples = apuSamples[:0]
			g.sync.addSamples(samples)
			if dropped {
				slog.Debug("dropped frame")
			}
		}
	}()

	return g, nil
}

type audioStream struct {
	sync *synchronization
}

func (w *audioStream) Read(p []byte) (n int, err error) {
	n = w.sync.waitConsumeSamples(p)
	return
}
