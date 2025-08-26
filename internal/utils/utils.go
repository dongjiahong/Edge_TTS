package utils

import (
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

// GenerateTextHash 生成文本哈希用于缓存
func GenerateTextHash(text, voice, format string) string {
	combined := fmt.Sprintf("%s|%s|%s", text, voice, format)
	hash := md5.Sum([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// GenerateRequestID 生成请求ID
func GenerateRequestID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenerateSSML 生成SSML格式的文本
func GenerateSSML(text, voice string, speed float64, pitch int) string {
	// 处理语音参数
	speedStr := fmt.Sprintf("%.1f", speed)
	if speed == 1.0 {
		speedStr = "default"
	}

	pitchStr := "default"
	if pitch != 0 {
		if pitch > 0 {
			pitchStr = fmt.Sprintf("+%dHz", pitch)
		} else {
			pitchStr = fmt.Sprintf("%dHz", pitch)
		}
	}

	ssml := fmt.Sprintf(`<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US">
		<voice name="%s">
			<prosody rate="%s" pitch="%s">
				%s
			</prosody>
		</voice>
	</speak>`, voice, speedStr, pitchStr, text)

	return ssml
}

// SanitizeFileName 清理文件名
func SanitizeFileName(name string) string {
	// 替换不安全的字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

// GetFileExtension 根据格式获取文件扩展名
func GetFileExtension(format string) string {
	switch strings.ToLower(format) {
	case "mp3":
		return ".mp3"
	case "wav":
		return ".wav"
	case "ogg":
		return ".ogg"
	case "m4a":
		return ".m4a"
	case "flac":
		return ".flac"
	default:
		return ".mp3"
	}
}