package main

import (
	"bytes"
	"regexp"

	"github.com/williamfhe/godivert"
	"github.com/williamfhe/godivert/header"
)

func rewrite(pkt *godivert.Packet, cfg *Config) (bool, bool) {
	pkt.ParseHeaders()
	if pkt.IpVersion() != 4 {
		return false, false
	}

	ipHdr, ok := pkt.IpHdr.(*header.IPv4Header)
	if !ok {
		return false, false
	}

	ttlChanged := false
	oldTTL := ipHdr.TTL()
	if oldTTL != cfg.TTL {
		pkt.Raw[8] = cfg.TTL
		adjustIPChecksum(pkt.Raw[:int(ipHdr.HeaderLen())], oldTTL, cfg.TTL)
		ttlChanged = true
	}

	if pkt.NextHeaderType() != header.TCP {
		return ttlChanged, false
	}

	tcp, ok := pkt.NextHeader.(*header.TCPHeader)
	if !ok {
		return ttlChanged, false
	}

	dstPort, err := tcp.DstPort()
	if err != nil || !containsPort(cfg.UA_Ports, dstPort) {
		return ttlChanged, false
	}

	ipHdrLen := int(ipHdr.HeaderLen())
	tcpHdrLen := tcp.HeaderLen()
	payloadStart := ipHdrLen + tcpHdrLen
	if payloadStart >= len(pkt.Raw) {
		return ttlChanged, false
	}

	payload := pkt.Raw[payloadStart:]
	if !looksLikeHTTPRequest(payload) {
		return ttlChanged, false
	}

	newPayload := replaceUserAgent(payload, cfg.UA)
	if newPayload == nil {
		return ttlChanged, false
	}

	if len(newPayload) == len(payload) {
		copy(pkt.Raw[payloadStart:], newPayload)
		return ttlChanged, true
	}

	newRaw := make([]byte, 0, payloadStart+len(newPayload))
	newRaw = append(newRaw, pkt.Raw[:payloadStart]...)
	newRaw = append(newRaw, newPayload...)
	pkt.Raw = newRaw
	pkt.PacketLen = uint(len(newRaw))

	newTotalLen := uint16(len(newRaw))
	pkt.Raw[2] = byte(newTotalLen >> 8)
	pkt.Raw[3] = byte(newTotalLen)
	recomputeIPChecksum(pkt.Raw[:ipHdrLen])

	return ttlChanged, true
}

func containsPort(ports []uint16, port uint16) bool {
	for _, candidate := range ports {
		if candidate == port {
			return true
		}
	}
	return false
}

func looksLikeHTTPRequest(payload []byte) bool {
	if len(payload) < 4 {
		return false
	}

	methods := [][]byte{
		[]byte("GET "),
		[]byte("POST "),
		[]byte("HEAD "),
		[]byte("PUT "),
		[]byte("DELETE "),
		[]byte("OPTIONS "),
		[]byte("PATCH "),
		[]byte("CONNECT "),
	}
	for _, method := range methods {
		if bytes.HasPrefix(payload, method) {
			return true
		}
	}
	return false
}

var uaRegex = regexp.MustCompile(`(?i)User-Agent:[^\r\n]*`)

func replaceUserAgent(payload []byte, newUA string) []byte {
	match := uaRegex.FindIndex(payload)
	if match != nil {
		out := make([]byte, len(payload))
		copy(out, payload)
		headerPrefix := []byte("User-Agent: ")
		lineLen := match[1] - match[0]
		valueLen := lineLen - len(headerPrefix)
		if valueLen <= 0 {
			return nil
		}
		value := []byte(newUA)
		if len(value) > valueLen {
			value = value[:valueLen]
		}
		copy(out[match[0]:match[1]], headerPrefix)
		valueStart := match[0] + len(headerPrefix)
		copy(out[valueStart:valueStart+valueLen], bytes.Repeat([]byte(" "), valueLen))
		copy(out[valueStart:valueStart+len(value)], value)
		return out
	}

	return nil
}

func adjustIPChecksum(hdr []byte, oldTTL, newTTL uint8) {
	if len(hdr) < 12 {
		return
	}

	oldWord := uint16(oldTTL) << 8
	newWord := uint16(newTTL) << 8
	checksum := uint32(hdr[10])<<8 | uint32(hdr[11])
	checksum += uint32(oldWord)
	checksum += ^uint32(newWord) & 0xFFFF

	for checksum > 0xFFFF {
		checksum = (checksum & 0xFFFF) + (checksum >> 16)
	}

	hdr[10] = byte(checksum >> 8)
	hdr[11] = byte(checksum)
}

func recomputeIPChecksum(hdr []byte) uint16 {
	if len(hdr) < 20 {
		return 0
	}

	hdr[10] = 0
	hdr[11] = 0
	sum := internetChecksum(hdr)
	hdr[10] = byte(sum >> 8)
	hdr[11] = byte(sum)
	return sum
}

func internetChecksum(data []byte) uint16 {
	var sum uint32

	i := 0
	for ; i+1 < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if i < len(data) {
		sum += uint32(data[i]) << 8
	}

	for sum > 0xFFFF {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return ^uint16(sum)
}
