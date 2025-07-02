package algorithms

import (
	"math"
)

const (
	// DefaultEmbeddingDimension 默认向量维度
	DefaultEmbeddingDimension = 128
)

// EmbeddingConfig 向量生成配置
type EmbeddingConfig struct {
	Dimension int  `json:"dimension"`
	Normalize bool `json:"normalize"`
	Seed      int  `json:"seed"`
}

// DefaultEmbeddingConfig 默认向量配置
func DefaultEmbeddingConfig() *EmbeddingConfig {
	return &EmbeddingConfig{
		Dimension: DefaultEmbeddingDimension,
		Normalize: true,
		Seed:      42,
	}
}

// GenerateEmbedding 生成文本向量嵌入
// 使用基于哈希的简化方法，适合零依赖环境
func GenerateEmbedding(text string, config *EmbeddingConfig) []float64 {
	if config == nil {
		config = DefaultEmbeddingConfig()
	}

	tokens := Tokenize(text)
	embedding := make([]float64, config.Dimension)

	// 基于词汇哈希生成向量
	for i, word := range tokens.Words {
		hash := HashString(word)
		for j := 0; j < config.Dimension; j++ {
			// 使用不同的哈希种子为每个维度生成值
			value := float64((hash+i*j+config.Seed)%100) / 100.0
			embedding[j] += value
		}
	}

	// L2归一化
	if config.Normalize {
		embedding = normalizeVector(embedding)
	}

	return embedding
}

// CosineSimilarity 计算余弦相似度
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	return dotProduct / (normA*normB + 1e-10)
}

// EuclideanDistance 计算欧几里得距离
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return math.Inf(1)
	}

	var sum float64
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// DotProduct 计算点积
func DotProduct(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var result float64
	for i := 0; i < len(a); i++ {
		result += a[i] * b[i]
	}

	return result
}

// normalizeVector L2归一化向量
func normalizeVector(vector []float64) []float64 {
	norm := 0.0
	for _, v := range vector {
		norm += v * v
	}

	if norm == 0 {
		return vector
	}

	norm = math.Sqrt(norm)
	normalized := make([]float64, len(vector))
	for i, v := range vector {
		normalized[i] = v / norm
	}

	return normalized
}

// VectorMagnitude 计算向量模长
func VectorMagnitude(vector []float64) float64 {
	var sum float64
	for _, v := range vector {
		sum += v * v
	}
	return math.Sqrt(sum)
}

// IsValidVector 检查向量是否有效
func IsValidVector(vector []float64) bool {
	if len(vector) == 0 {
		return false
	}

	for _, v := range vector {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}

	return true
}
