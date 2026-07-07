package godivert

import (
	"encoding/binary"
	"fmt"
)

// Represents a WinDivertAddress struct
// See : https://reqrypt.org/windivert-doc.html#divert_address
// As go doesn't not support bit fields
// we use a little trick to get the Direction, Loopback, Import and PseudoChecksum fields
type WinDivertAddress struct {
	Timestamp int64
	Flags     uint32
	Reserved1 uint32
	Data      [64]byte
}

func (w *WinDivertAddress) String() string {
	return fmt.Sprintf("{\n"+
		"\t\tTimestamp=%d\n"+
		"\t\tInteface={IfIdx=%d SubIfIdx=%d}\n"+
		"\t\tDirection=%v\n"+
		"\t\tLoopback=%t\n"+
		"\t\tImpostor=%t\n"+
		"\t\tPseudoChecksum={IP=%t TCP=%t UDP=%t}\n"+
		"\t}",
		w.Timestamp, w.IfIdx(), w.SubIfIdx(), w.Direction(), w.Loopback(), w.Impostor(),
		w.PseudoIPChecksum(), w.PseudoTCPChecksum(), w.PseudoUDPChecksum())
}

func (w *WinDivertAddress) IfIdx() uint32 {
	return binary.LittleEndian.Uint32(w.Data[0:4])
}

func (w *WinDivertAddress) SubIfIdx() uint32 {
	return binary.LittleEndian.Uint32(w.Data[4:8])
}

// Returns the direction of the packet
// WinDivertDirectionInbound (true) for inbounds packets
// WinDivertDirectionOutbounds (false) for outbounds packets
func (w *WinDivertAddress) Direction() Direction {
	return Direction((w.Flags & 0x1) == 1)
}

// Returns true if the packet is a loopback packet
func (w *WinDivertAddress) Loopback() bool {
	return ((w.Flags >> 1) & 0x1) == 1
}

// Returns true if the packet is an impostor
// See https://reqrypt.org/windivert-doc.html#divert_address for more information
func (w *WinDivertAddress) Impostor() bool {
	return ((w.Flags >> 2) & 0x1) == 1
}

// Returns true if the packet uses a pseudo IP checksum
func (w *WinDivertAddress) PseudoIPChecksum() bool {
	return ((w.Flags >> 3) & 0x1) == 1
}

// Returns true if the packet uses a pseudo TCP checksum
func (w *WinDivertAddress) PseudoTCPChecksum() bool {
	return ((w.Flags >> 4) & 0x1) == 1
}

// Returns true if the packet uses a pseudo UDP checksum
func (w *WinDivertAddress) PseudoUDPChecksum() bool {
	return ((w.Flags >> 5) & 0x1) == 1
}
