package ppu

import "fmt"

type DMAMode int

const (
	DMAIdle  DMAMode = iota
	DMAInit1         // M-cycle in which the write to FF46 happens. OAM accessible
	DMAInit2         // M-cycle after write to FF46. OAM accessible
	// Active transfer.
	DMATransfer // OAM inaccessible
	// M-cycle in which the write to FF46 happens but transfer is still happening.
	DMAInitTransferring1 // OAM *inacessible*, bytes are transferred.
	// 2nd M-cycle of above state. Next state is DMATransfer.
	DMAInitTransferring2 // OAM *inaccessible*, bytes are transferred.
)

type DMAUnit struct {
	// TODO: move clock outside unit
	clock                uint8 // runs 0, 1, 2, 3 every T-cycle
	addressReg           uint8
	addressHi, addressLo uint8

	State DMAMode
}

func (dma *DMAUnit) WriteRegister(value uint8) {
	dma.addressReg = value

	switch dma.State {
	case DMAIdle, DMAInit1, DMAInit2:
		dma.State = DMAInit1
	case DMATransfer, DMAInitTransferring1, DMAInitTransferring2:
		dma.State = DMAInitTransferring1
	default:
		panic(fmt.Sprintf("unexpected dma state %v", dma.State))
	}
}

// StartCycle performs the work for the DMA at the start of a cycle.
// It returns (active, address) = (external bus read needed, bus address)
func (dma DMAUnit) StartCycle() (active bool, address uint16) {
	switch dma.State {
	case DMAInitTransferring1, DMAInitTransferring2, DMATransfer:
		active = true
		address = uint16(dma.addressHi)<<8 | uint16(dma.addressLo)
	}

	return
}

func (dma *DMAUnit) stepClock() bool {
	dma.clock = (dma.clock + 1) & 3
	return dma.clock == 0
}

// Step DMA unit's logic for T-cycle.
func (dma *DMAUnit) Step(oam *OAMMem, data byte) {
	if !dma.stepClock() {
		return
	}

	active := dma.State >= DMATransfer && int(dma.addressLo) < len(oam)

	if active {
		oam[dma.addressLo] = data
		dma.addressLo++
		active = int(dma.addressLo) < len(oam)
	}

	// advance state
	switch dma.State {
	case DMATransfer:
		if !active {
			*dma = DMAUnit{}
		}
	case DMAInit1, DMAInitTransferring1:
		dma.State++
	case DMAInit2, DMAInitTransferring2:
		dma.State = DMATransfer
		dma.addressHi, dma.addressLo = dma.addressReg, 0
	}
}
