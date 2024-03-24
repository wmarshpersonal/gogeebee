package gb

import "sync"

type BusOperation int

const (
	RD BusOperation = iota + 1
	WR
)

type bank int

const (
	main bank = iota
	hram
)

type Token struct {
	bank
	valid bool
}

type BusSelect struct {
	m        sync.Mutex
	requests [2]struct {
		once sync.Once

		op   BusOperation
		addr uint16
		data byte
	}
	results [2]byte
}

func (bs *BusSelect) Select(op BusOperation, address uint16, data byte) Token {
	var b bank
	if address >= 0xFF80 && address <= 0xFFEE {
		b = hram
	}

	bs.m.Lock()
	bs.requests[b].op = op
	bs.requests[b].addr = address
	bs.requests[b].data = data
	bs.m.Unlock()

	return Token{b, true}
}

func (bs *BusSelect) SelectRead(address uint16) Token {
	return bs.Select(RD, address, 0)
}

func (bs *BusSelect) SelectWrite(address uint16, data byte) Token {
	return bs.Select(WR, address, data)
}

func (bs *BusSelect) Commit(
	token Token, read BusRead, write BusWrite,
) (data byte) {
	if !token.valid {
		return
	}

	req := &bs.requests[token.bank]
	req.once.Do(func() {
		op, addr, data := req.op, req.addr, req.data
		switch op {
		case RD:
			data = read(addr)
		case WR:
			write(addr, data)
		}
		bs.results[token.bank] = data
	})

	return bs.results[token.bank]
}
