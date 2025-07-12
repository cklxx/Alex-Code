package utils

import (
	"math/rand"
	"time"
)

// Random processing status messages list (without dots)
var processingMessages = []string{
	"Processing",
	"Thinking",
	"Learning",
	"Exploring",
	"Discovering",
	"Analyzing",
	"Computing",
	"Reasoning",
	"Planning",
	"Executing",
	"Optimizing",
	"Searching",
	"Understanding",
	"Crafting",
	"Creating",
	"Parsing",
	"Generating",
	"Evaluating",
	"Calculating",
	"Investigating",
	"Synthesizing",
	"Interpreting",
	"Assembling",
	"Constructing",
	"Formulating",
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRandomProcessingMessage 获取随机处理消息
func GetRandomProcessingMessage() string {
	return processingMessages[rng.Intn(len(processingMessages))] + "..."
}

// GetRandomProcessingMessageWithEmoji 获取带emoji的随机处理消息
func GetRandomProcessingMessageWithEmoji() string {
	return "⚡ " + GetRandomProcessingMessage() + " please wait"
}