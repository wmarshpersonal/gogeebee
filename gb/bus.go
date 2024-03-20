package gb

type Bus interface {
	Read(address uint16) (value uint8)
	Write(address uint16, value uint8)
}

// DevBus is an unoptimized convenience bus implementation
// designed for development only.
type DevBus struct {
	slots [0x10000]*busSlot
}

type busSlot struct {
	r func(address uint16) (value uint8)
	w func(address uint16, value uint8)
}

func NewDevBus() *DevBus {
	return &DevBus{}
}

// Connect hooks something up to the bus by letting it supply
// its read/write functions and an address (or range of addresses)
// it's accessible at.
// Connect will panic if the connection overlaps with something
// previously connected or if either read/write are nil.
func (b *DevBus) Connect(
	read func(address uint16) (value uint8),
	write func(address uint16, value uint8),
	address uint16,
) {
	if b.slots[address] != nil {
		panic("slot overlap")
	}
	b.slots[address] = &busSlot{read, write}
}

// ConnectRange is like Connect, but fills multiple slots in an
// inclusive range [start, end].
// panics if end < start. end == start is the same as a Connect
// with a single address.
func (b *DevBus) ConnectRange(
	read func(address uint16) (value uint8),
	write func(address uint16, value uint8),
	start, end uint16,
) {
	if end < start {
		panic("end < start")
	}

	for i := start; i <= end; i++ {
		b.Connect(read, write, i)
	}
}

// Read will panic if the slot at address is empty.
func (b *DevBus) Read(address uint16) uint8 {
	slot := b.slots[address]
	if slot == nil {
		panic("nil slot")
	}
	return slot.r(address)
}

// Write will panic if the slot at address is empty.
func (b *DevBus) Write(address uint16, value uint8) {
	slot := b.slots[address]
	if slot == nil {
		panic("nil slot")
	}
	slot.w(address, value)
}
