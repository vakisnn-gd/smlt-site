package main

import (
	"strings"
	"testing"
)

func TestDiscordNotificationKeyIsPostgresSafe(t *testing.T) {
	key := discordNotificationKey("demon", "текст\x00с нулевым байтом")

	if len(key) != 64 {
		t.Fatalf("discord notification key length = %d, want 64", len(key))
	}
	if strings.IndexByte(key, 0) >= 0 {
		t.Fatal("discord notification key contains a NUL byte")
	}
	if key == discordNotificationKey("rank", "текст\x00с нулевым байтом") {
		t.Fatal("different Discord event types produced the same key")
	}
}
