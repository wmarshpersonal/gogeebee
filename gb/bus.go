package gb

// Bus state encoding with (W)rite, (R)ead, (C)hip select, (A)ddress, (D)ata:
// FEDCBA9876543210FEDCBA9876543210
// #####WRCAAAAAAAAAAAAAAAADDDDDDDD
type ExternalBus uint32

// FEDCBA9876543210FEDCBA9876543210
// ########WRCAAAAAAAAAAAAADDDDDDDD
type VideoBus uint32

// type BusSelection int

// const (
// 	Ext BusSelection = iota + 1
// 	Vid
// )

// func selectBus(addr uint16) BusSelection {
// 	if addr>>13 == 4 { // 0x8000-0x9FFF
// 		return Vid
// 	}
// 	return Ext
// }

type DMGBus struct {
	Ext ExternalBus
	Vid VideoBus
}
