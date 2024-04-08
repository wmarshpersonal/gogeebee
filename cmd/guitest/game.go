package main

import (
	"log/slog"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/gb"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

type Game struct {
	gb   *gb.GB
	sync synchronization
}

func (g *Game) Update() error {
	// g.gb.ProcessJoypad(ReadButtons(), ReadDirections())

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

	var fa atomic.Value
	fa.Store(440.0)

	go func() {
		for {
			f := 440.
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				f = 520.
			}
			fa.Store(f)
			runtime.Gosched()
		}
	}()

	go func() {
		var phase float64
		var samples []byte
		var frame ppu.PixelBuffer
		for {
			drawn := g.gb.RunFor(gb.TCyclesPerSecond*(bufferSize/8)/sampleRate, &frame)
			var dropped bool
			if drawn > 0 {
				dropped = g.sync.addFrame(frame)
			}
			samples = samples[:0]
			const samplesPerFrame = bufferSize / 4 / chs
			f := fa.Load().(float64)
			for i := 0; i < samplesPerFrame; i++ {
				phase += f * 2 * math.Pi / sampleRate
				bits := math.Float32bits(0.3 * float32(math.Sin(phase)))
				samples = append(samples, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))
				samples = append(samples, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))
			}
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
