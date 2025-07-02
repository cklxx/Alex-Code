package context

import (
	"context"

	"alex/internal/context/algorithms"
)

// === 便利函数 - 使用统一引擎 ===

var defaultEngine *UnifiedEngine

// init 初始化默认引擎
func init() {
	defaultEngine = NewUnifiedEngine(DefaultEngineConfig())
}

// EnhancePrompt 增强现有提示的工具函数
func EnhancePrompt(prompt, task, input string) string {
	ctx := context.Background()
	contextResult, err := defaultEngine.BuildContext(ctx, task, input)
	if err != nil {
		return prompt // 失败时返回原始提示
	}

	return prompt + "\n\nContext:\n" + contextResult.Content
}

// CompressText 文本压缩工具函数
func CompressText(text string, ratio float64) string {
	return algorithms.CompressText(text, ratio)
}

// SearchRelevant 检索相关信息工具函数
func SearchRelevant(query string, limit int) []string {
	results, err := defaultEngine.SearchSimilar(query, limit)
	if err != nil {
		return []string{}
	}

	var relevantInfo []string
	for _, result := range results {
		relevantInfo = append(relevantInfo, result.Document.Title+": "+result.Document.Content)
	}

	return relevantInfo
}

// AddDocument 添加文档到默认引擎
func AddDocument(doc Document) error {
	return defaultEngine.AddDocument(doc)
}

// SearchSimilar 向量相似搜索工具函数
func SearchSimilar(query string, limit int) ([]VectorResult, error) {
	return defaultEngine.SearchSimilar(query, limit)
}

// === 公共算法函数导出 ===

// Tokenize 分词工具函数
func Tokenize(text string) []string {
	result := algorithms.Tokenize(text)
	return result.Words
}

// GenerateEmbedding 生成向量嵌入
func GenerateEmbedding(text string) []float64 {
	return algorithms.GenerateEmbedding(text, algorithms.DefaultEmbeddingConfig())
}

// CosineSimilarity 余弦相似度计算
func CosineSimilarity(a, b []float64) float64 {
	return algorithms.CosineSimilarity(a, b)
}

// CalculateQuality 计算质量分数
func CalculateQuality(content string, contextCount int) float64 {
	metrics := algorithms.CalculateQuality(content, contextCount)
	return metrics.FinalScore
}

// HashString 字符串哈希函数
func HashString(s string) int {
	return algorithms.HashString(s)
}

// GetDefaultEngine 获取默认引擎（用于高级用法）
func GetDefaultEngine() *UnifiedEngine {
	return defaultEngine
}

// ResetDefaultEngine 重置默认引擎（主要用于测试）
func ResetDefaultEngine() {
	if defaultEngine != nil {
		defaultEngine.Close()
	}
	defaultEngine = NewUnifiedEngine(DefaultEngineConfig())
}
