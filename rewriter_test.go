package main

import (
	"bytes"
	"testing"
)

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
