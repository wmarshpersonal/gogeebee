package ppu

const OAMSize int = 0xA0

type DMAMode int

const (
	DMAIdle DMAMode = iota
	DMAStartup
	DMATransfer
)

type DMAUnit struct {
	i int

	Mode    DMAMode
	Address uint16
}

func (dma *DMAUnit) StepM(oam []byte, value byte) {
	switch dma.Mode {
	case DMAIdle:
	case DMAStartup:
		dma.Mode = DMATransfer
	case DMATransfer:
		oam[dma.i] = value
		dma.i++
		dma.Address++
		if dma.i == OAMSize { // reset
			*dma = DMAUnit{}
		}
	default:
		panic("dma mode")
	}
}
