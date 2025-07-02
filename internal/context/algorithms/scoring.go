package algorithms

import (
	"math"
	"strings"
)

// QualityMetrics 质量指标
type QualityMetrics struct {
	ContentLength     int     `json:"content_length"`
	ContextCount      int     `json:"context_count"`
	HasTaskInput      bool    `json:"has_task_input"`
	StructureScore    float64 `json:"structure_score"`
	CompletenessScore float64 `json:"completeness_score"`
	FinalScore        float64 `json:"final_score"`
}

// CalculateQuality 计算内容质量分数
func CalculateQuality(content string, contextCount int) *QualityMetrics {
	metrics := &QualityMetrics{
		ContentLength: len(content),
		ContextCount:  contextCount,
	}

	if content == "" {
		return metrics
	}

	// 基础分数
	score := 0.5

	// 内容长度评分 (30%)
	if len(content) > 50 {
		score += 0.3
	} else if len(content) > 20 {
		score += 0.15
	}

	// 结构完整性评分 (20%)
	hasTask := strings.Contains(content, "Task:")
	hasInput := strings.Contains(content, "Input:")
	metrics.HasTaskInput = hasTask && hasInput

	if metrics.HasTaskInput {
		score += 0.2
		metrics.StructureScore = 1.0
	} else if hasTask || hasInput {
		score += 0.1
		metrics.StructureScore = 0.5
	}

	// 上下文丰富度评分 (20%)
	if contextCount > 0 {
		contextScore := 0.2 * float64(contextCount) / 5.0 // 最多5个上下文
		if contextScore > 0.2 {
			contextScore = 0.2
		}
		score += contextScore
		metrics.CompletenessScore = contextScore / 0.2
	}

	// 确保分数在合理范围内
	if score > 1.0 {
		score = 1.0
	}

	metrics.FinalScore = score
	return metrics
}

// ScoreResult 搜索结果评分结构
type ScoreResult struct {
	IndexScore  float64 `json:"index_score"`
	VectorScore float64 `json:"vector_score"`
	TextScore   float64 `json:"text_score"`
	FinalScore  float64 `json:"final_score"`
	Explanation string  `json:"explanation"`
}

// CalculateHybridScore 计算混合搜索评分
func CalculateHybridScore(indexScore, vectorScore, textMatchScore float64) *ScoreResult {
	// 权重配置
	const (
		indexWeight  = 0.4 // 倒排索引权重
		vectorWeight = 0.4 // 向量相似度权重
		textWeight   = 0.2 // 文本匹配权重
	)

	finalScore := indexScore*indexWeight + vectorScore*vectorWeight + textMatchScore*textWeight

	var explanation strings.Builder
	explanation.WriteString("Hybrid scoring: ")
	if indexScore > 0 {
		explanation.WriteString("keyword match ")
	}
	if vectorScore > 0 {
		explanation.WriteString("semantic similarity ")
	}
	if textMatchScore > 0 {
		explanation.WriteString("text overlap ")
	}

	return &ScoreResult{
		IndexScore:  indexScore,
		VectorScore: vectorScore,
		TextScore:   textMatchScore,
		FinalScore:  finalScore,
		Explanation: explanation.String(),
	}
}

// TFIDFScore 计算TF-IDF评分
func TFIDFScore(termFreq, docFreq, totalDocs int) float64 {
	if termFreq == 0 || docFreq == 0 || totalDocs == 0 {
		return 0.0
	}

	// TF计算 (词频)
	tf := float64(termFreq)

	// IDF计算 (逆文档频率)
	idf := math.Log(float64(totalDocs) / float64(docFreq))

	return tf * idf
}

// RelevanceThreshold 相关性阈值配置
type RelevanceThreshold struct {
	MinScore       float64 `json:"min_score"`
	HighScore      float64 `json:"high_score"`
	ExcellentScore float64 `json:"excellent_score"`
}

// DefaultRelevanceThreshold 默认相关性阈值
func DefaultRelevanceThreshold() *RelevanceThreshold {
	return &RelevanceThreshold{
		MinScore:       0.05,
		HighScore:      0.3,
		ExcellentScore: 0.7,
	}
}

// ClassifyRelevance 分类相关性等级
func ClassifyRelevance(score float64, threshold *RelevanceThreshold) string {
	if threshold == nil {
		threshold = DefaultRelevanceThreshold()
	}

	switch {
	case score >= threshold.ExcellentScore:
		return "excellent"
	case score >= threshold.HighScore:
		return "high"
	case score >= threshold.MinScore:
		return "moderate"
	default:
		return "low"
	}
}
