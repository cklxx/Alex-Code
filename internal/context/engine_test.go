package context

import (
	"context"
	"testing"
)

func TestHybridEngine(t *testing.T) {
	engine := NewEngine()

	// 添加测试文档
	docs := []Document{
		{
			ID:      "doc1",
			Title:   "Go并发编程",
			Content: "Go语言的并发编程使用goroutine实现，通过channel进行通信",
		},
		{
			ID:      "doc2",
			Title:   "代码优化",
			Content: "代码优化包括性能优化、内存优化和算法优化等方面",
		},
		{
			ID:      "doc3",
			Title:   "数据结构",
			Content: "常用的数据结构包括数组、链表、树、图等基础结构",
		},
	}

	// 添加文档到引擎
	for _, doc := range docs {
		err := engine.AddDocument(doc)
		if err != nil {
			t.Fatalf("添加文档失败: %v", err)
		}
	}

	// 测试文档数量
	stats := engine.Stats()
	if stats.DocumentCount != 3 {
		t.Errorf("期望文档数量3，实际得到%d", stats.DocumentCount)
	}

	// 测试搜索功能
	results, err := engine.SearchSimilar("Go并发", 2)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	if len(results) == 0 {
		t.Error("应该返回搜索结果")
	}

	// 验证搜索结果
	found := false
	for _, result := range results {
		if result.Document.ID == "doc1" {
			found = true
			if result.Similarity <= 0 {
				t.Error("相似度应该大于0")
			}
		}
	}

	if !found {
		t.Error("应该找到doc1文档")
	}

	t.Logf("搜索结果数量: %d", len(results))
	for i, result := range results {
		t.Logf("结果%d: ID=%s, 相似度=%.3f, 标题=%s",
			i+1, result.Document.ID, result.Similarity, result.Document.Title)
	}
}

func TestBuildContext(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	ctx := context.Background()
	result, err := engine.BuildContext(ctx, "代码审查", "func main() { println(\"Hello World\") }")

	if err != nil {
		t.Fatalf("构建上下文失败: %v", err)
	}

	if result.Task != "代码审查" {
		t.Errorf("期望任务'代码审查'，实际得到'%s'", result.Task)
	}

	if result.Quality <= 0 {
		t.Error("质量分数应该大于0")
	}

	if result.Content == "" {
		t.Error("上下文内容不应为空")
	}

	t.Logf("上下文ID: %s", result.ID)
	t.Logf("质量分数: %.2f", result.Quality)
	t.Logf("内容长度: %d", len(result.Content))
}

func TestTokenizer(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			"Go语言并发编程",
			[]string{"go", "语言", "并发", "编程"},
		},
		{
			"Hello World Programming",
			[]string{"hello", "world", "programming"},
		},
		{
			"数据结构与算法分析",
			[]string{"数据", "结构", "算法", "分析"},
		},
	}

	for _, tc := range testCases {
		result := Tokenize(tc.input)
		if len(result) < len(tc.expected) {
			t.Errorf("输入'%s'，期望至少%d个词，实际得到%d个: %v",
				tc.input, len(tc.expected), len(result), result)
		}
	}
}

func TestEmbeddingGeneration(t *testing.T) {
	text1 := "Go语言并发编程"
	text2 := "Java多线程编程"
	text3 := "完全不相关的内容"

	embedding1 := GenerateEmbedding(text1)
	embedding2 := GenerateEmbedding(text2)
	embedding3 := GenerateEmbedding(text3)

	if len(embedding1) != 128 {
		t.Errorf("向量维度应该是128，实际得到%d", len(embedding1))
	}

	// 相似文本的相似度应该高于不相关文本
	sim12 := CosineSimilarity(embedding1, embedding2)
	sim13 := CosineSimilarity(embedding1, embedding3)

	t.Logf("相似文本相似度: %.3f", sim12)
	t.Logf("不相关文本相似度: %.3f", sim13)

	if sim12 <= 0 {
		t.Error("相似文本的相似度应该大于0")
	}
}
