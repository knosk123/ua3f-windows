package godivert

type Direction bool

const (
	// WinDivert 2.x may return packets larger than the Ethernet MTU. Its
	// official WINDIVERT_MTU_MAX is 40 + 0xFFFF.
	PacketBufferSize   = 0xFFFF + 40
	PacketChanCapacity = 256

	WinDivertDirectionOutbound Direction = false
	WinDivertDirectionInbound  Direction = true
)

const (
	WinDivertFlagSniff uint8 = 1 << iota
	WinDivertFlagDrop  uint8 = 1 << iota
	WinDivertFlagDebug uint8 = 1 << iota
)

func (d Direction) String() string {
	if bool(d) {
		return "Inbound"
	}
	return "Outbound"
}
