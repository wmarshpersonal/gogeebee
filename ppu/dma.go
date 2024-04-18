package ppu

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

func (dma *DMAUnit) StepM(oam *OAMMem, value byte) {
	switch dma.Mode {
	case DMAIdle:
	case DMAStartup:
		dma.Mode = DMATransfer
	case DMATransfer:
		oam[dma.i] = value
		dma.i++
		dma.Address++
		if dma.i == len(oam) { // reset
			*dma = DMAUnit{}
		}
	default:
		panic("dma mode")
	}
}
