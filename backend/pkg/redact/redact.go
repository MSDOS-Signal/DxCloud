// Package redact：敏感字段脱敏（审计/日志通用）。
package redact

import "strings"

var sensitiveWords = []string{
	"password", "passwd", "pwd", "token", "secret", "authorization",
	"api_key", "apikey", "access_key", "private_key", "credential",
}

func isSensitiveKey(k string) bool {
	lower := strings.ToLower(k)
	for _, w := range sensitiveWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

// Map 递归脱敏：键名命中敏感词的字段替换为 "***"，其余保留（嵌套 map 递归处理）。
func Map(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			out[k] = "***"
			continue
		}
		if sub, ok := v.(map[string]any); ok {
			out[k] = Map(sub)
			continue
		}
		if items, ok := v.([]map[string]any); ok {
			subOut := make([]map[string]any, 0, len(items))
			for _, it := range items {
				subOut = append(subOut, Map(it))
			}
			out[k] = subOut
			continue
		}
		out[k] = v
	}
	return out
}
