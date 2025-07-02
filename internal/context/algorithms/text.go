package algorithms

import (
	"strings"
)

// TokenizeResult 分词结果
type TokenizeResult struct {
	Words     []string `json:"words"`
	StopWords int      `json:"stop_words_removed"`
}

// stopWords 中英文停用词表
var stopWords = map[string]bool{
	// 英文停用词
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true,
	"by": true, "for": true, "from": true, "has": true, "he": true, "in": true, "is": true,
	"it": true, "its": true, "of": true, "on": true, "that": true, "the": true, "to": true,
	"was": true, "were": true, "will": true, "with": true,
	// 中文停用词
	"这": true, "是": true, "的": true, "了": true, "和": true, "在": true, "有": true,
	"一": true, "个": true, "中": true,
}

// Tokenize 智能分词函数 - 支持中英文混合文本
func Tokenize(text string) *TokenizeResult {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return &TokenizeResult{Words: []string{}, StopWords: 0}
	}

	var result []string
	var currentWord []rune
	stopWordsCount := 0

	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			// 英文字符和数字
			currentWord = append(currentWord, r)
		} else if r >= 0x4e00 && r <= 0x9fff {
			// 中文字符
			if len(currentWord) > 0 {
				word := string(currentWord)
				if len(word) > 1 && !stopWords[word] {
					result = append(result, word)
				} else if stopWords[word] {
					stopWordsCount++
				}
				currentWord = nil
			}
			char := string(r)
			if !stopWords[char] {
				result = append(result, char)
			} else {
				stopWordsCount++
			}
		} else {
			// 分隔符
			if len(currentWord) > 0 {
				word := string(currentWord)
				if len(word) > 1 && !stopWords[word] {
					result = append(result, word)
				} else if stopWords[word] {
					stopWordsCount++
				}
				currentWord = nil
			}
		}
	}

	// 处理最后一个词
	if len(currentWord) > 0 {
		word := string(currentWord)
		if len(word) > 1 && !stopWords[word] {
			result = append(result, word)
		} else if stopWords[word] {
			stopWordsCount++
		}
	}

	return &TokenizeResult{
		Words:     result,
		StopWords: stopWordsCount,
	}
}

// CompressText 智能文本压缩
func CompressText(text string, ratio float64) string {
	if ratio <= 0 || ratio >= 1 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	targetLength := int(float64(len(words)) * ratio)
	if targetLength >= len(words) || targetLength < 5 {
		return text
	}

	// 简单截取策略 - 可以后续优化为保留重要词汇
	return strings.Join(words[:targetLength], " ")
}

// CalculateTextMatch 计算文本匹配度
func CalculateTextMatch(query, content string) float64 {
	query = strings.ToLower(query)
	content = strings.ToLower(content)

	// 完全包含得分最高
	if strings.Contains(content, query) {
		return 1.0
	}

	// 计算词汇重叠度
	queryTokens := Tokenize(query)
	contentTokens := Tokenize(content)

	if len(queryTokens.Words) == 0 {
		return 0.0
	}

	matches := 0
	for _, qw := range queryTokens.Words {
		for _, cw := range contentTokens.Words {
			if qw == cw {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(queryTokens.Words))
}

// HashString 字符串哈希函数
func HashString(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}
