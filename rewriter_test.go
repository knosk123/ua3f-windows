package main

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/williamfhe/godivert"
)

func TestRewriteUserAgentOnAnyTCPPort(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nUser-Agent: original-agent/0.1\r\n\r\n")
	pkt := makeIPv4TCPPacket(43210, payload)
	cfg := &Config{TTL: 64, UA: "mobile"}

	_, uaChanged := rewrite(pkt, cfg)

	if !uaChanged {
		t.Fatal("expected User-Agent rewrite on a non-standard TCP port")
	}
	if !bytes.Contains(pkt.Raw, []byte("User-Agent: mobile")) {
		t.Fatalf("rewritten packet does not contain target User-Agent: %q", pkt.Raw)
	}
}

func makeIPv4TCPPacket(dstPort uint16, payload []byte) *godivert.Packet {
	raw := make([]byte, 40+len(payload))
	raw[0] = 0x45
	binary.BigEndian.PutUint16(raw[2:4], uint16(len(raw)))
	raw[8] = 64
	raw[9] = 6
	binary.BigEndian.PutUint16(raw[20:22], 50000)
	binary.BigEndian.PutUint16(raw[22:24], dstPort)
	raw[32] = 0x50
	raw[33] = 0x18
	copy(raw[40:], payload)
	return &godivert.Packet{Raw: raw, PacketLen: uint(len(raw))}
}

func TestReplaceUserAgentReplacesExistingHeader(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nUser-Agent: old\r\n\r\n")
	got := replaceUserAgent(payload, "new-agent")
	want := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nUser-Agent: new\r\n\r\n")

	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected payload:\nwant=%q\ngot=%q", want, got)
	}
}

func TestReplaceUserAgentPadsShorterValueToPreservePayloadLength(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nUser-Agent: original-agent/0.1\r\n\r\n")
	got := replaceUserAgent(payload, "short")

	if len(got) != len(payload) {
		t.Fatalf("payload length changed: want=%d got=%d", len(payload), len(got))
	}
	if !bytes.Contains(got, []byte("User-Agent: short             \r\n")) {
		t.Fatalf("short UA was not padded in-place: %q", got)
	}
}

func TestReplaceUserAgentTruncatesLongerValueToPreservePayloadLength(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nUser-Agent: original-agent/0.1\r\n\r\n")
	got := replaceUserAgent(payload, "ua3f-agent-0000001-extra")

	if len(got) != len(payload) {
		t.Fatalf("payload length changed: want=%d got=%d", len(payload), len(got))
	}
	if !bytes.Contains(got, []byte("User-Agent: ua3f-agent-0000001\r\n")) {
		t.Fatalf("long UA was not truncated in-place: %q", got)
	}
}

func TestReplaceUserAgentDoesNotInsertMissingHeader(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nAccept: */*\r\n\r\n")
	got := replaceUserAgent(payload, "new-agent")

	if got != nil {
		t.Fatalf("missing User-Agent should not be inserted in stable mode: %q", got)
	}
}
