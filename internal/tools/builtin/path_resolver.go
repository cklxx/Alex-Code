package builtin

import (
	"context"
	"os"
	"path/filepath"
)

// ContextKey 用于在context中存储工作目录
type ContextKey string

const WorkingDirKey ContextKey = "working_dir"

// PathResolver 路径解析器
type PathResolver struct {
	workingDir string
}

// NewPathResolver 创建新的路径解析器
func NewPathResolver(workingDir string) *PathResolver {
	if workingDir == "" {
		workingDir, _ = os.Getwd()
	}
	return &PathResolver{workingDir: workingDir}
}

// ResolvePath 解析路径，将相对路径转换为绝对路径
func (pr *PathResolver) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	
	// 将相对路径基于工作目录解析为绝对路径
	resolved := filepath.Join(pr.workingDir, path)
	return filepath.Clean(resolved)
}

// GetPathResolverFromContext 从context获取路径解析器
func GetPathResolverFromContext(ctx context.Context) *PathResolver {
	if ctx == nil {
		return NewPathResolver("")
	}
	
	if workingDir, ok := ctx.Value(WorkingDirKey).(string); ok {
		return NewPathResolver(workingDir)
	}
	
	return NewPathResolver("")
}

// WithWorkingDir 在context中设置工作目录
func WithWorkingDir(ctx context.Context, workingDir string) context.Context {
	return context.WithValue(ctx, WorkingDirKey, workingDir)
}