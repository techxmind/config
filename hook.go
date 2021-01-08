package config

import (
	"encoding/json"
	"regexp"
)

var (
	jsonCommentRegexps = []*regexp.Regexp{
		regexp.MustCompile(`(?m)//[^"]+?$`),
		regexp.MustCompile(`(?m)^\s*//.*?$`),
	}
)

// 对原始配置内容进行处理，像解密加密文本等
type RawMessageProcessor func([]byte) []byte

var processors []RawMessageProcessor

func init() {
	processors = make([]RawMessageProcessor, 0, 1)
	RegisterRawMessageProcessor(trimJsonComment)
}

func RegisterRawMessageProcessor(p RawMessageProcessor) {
	processors = append(processors, p)
}

func processRawMessage(msg []byte) []byte {
	if len(msg) == 0 {
		return nil
	}

	ret := msg
	for _, p := range processors {
		ret = p(ret)
	}
	return ret
}

func trimJsonComment(content []byte) []byte {
	s := string(content)
	for _, r := range jsonCommentRegexps {
		s = r.ReplaceAllString(s, "")
	}

	if string(content) == s {
		return content
	}

	var v interface{}
	err := json.Unmarshal([]byte(s), &v)
	if err == nil {
		return []byte(s)
	}

	return content
}
