package util

import "strings"

// 转义字符映射
var escapeCharacters = map[string]rune{
	`\0`: 0,
	`\a`: 7,
	`\b`: 8,
	`\t`: 9,
	`\n`: 10,
	`\v`: 11,
	`\f`: 12,
	`\r`: 13,
	`\e`: 27,
	`\\`: 92,
}

// ParseEscapeCharacter 格式化转义字符
func ParseEscapeCharacter(s string, oldnew ...string) string {
	oldnews := make([]string, len(escapeCharacters)*2+len(oldnew))
	var index int
	for f, t := range escapeCharacters {
		oldnews[index] = f
		oldnews[index+1] = string([]rune{t})
		index += 2
	}
	for i, s := range oldnew {
		oldnews[index+i] = s
	}
	replacer := strings.NewReplacer(oldnews...)
	return replacer.Replace(s)
}
