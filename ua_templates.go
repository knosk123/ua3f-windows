package main

import "strings"

var presets = map[string]string{
	"pc":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:149.0) Gecko/20100101 Firefox/149.0",
	"wechat": "Mozilla/5.0 (Linux; Android 15; RMX6688 Build/AP3A.240617.008; wv) AppleWebKit/537.36",
}

// 默认 UA：wechat（手机/微信内嵌浏览器场景）
func defaultUA() string {
	if ua, ok := presets["wechat"]; ok {
		return ua
	}
	return "Mozilla/5.0"
}

// 解析 UA 参数：如果用户输入的是预设名（不带空格），返回对应 UA；否则原样返回
func resolveUA(input string) string {
	if ua, ok := presets[strings.ToLower(strings.TrimSpace(input))]; ok {
		return ua
	}
	return presets["wechat"]
}
