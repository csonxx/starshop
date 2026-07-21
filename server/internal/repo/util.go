// Package repo 私有辅助: 列表解析等.
package repo

import "strings"

// splitList 解析逗号分隔字符串, 去空去重.
func splitList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func SplitLists(values ...string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, value := range values {
		for _, item := range splitList(value) {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}
