package utils

import (
	"strings"
	"testing"
)

func TestGetRandomProcessingMessage(t *testing.T) {
	msg := GetRandomProcessingMessage()
	if msg == "" || !strings.HasSuffix(msg, "...") {
		t.Errorf("Invalid processing message: %s", msg)
	}
}

func TestGetRandomProcessingMessageWithEmoji(t *testing.T) {
	msg := GetRandomProcessingMessageWithEmoji()
	if !strings.HasPrefix(msg, "⚡ ") || !strings.HasSuffix(msg, " please wait") {
		t.Errorf("Invalid emoji processing message: %s", msg)
	}
}