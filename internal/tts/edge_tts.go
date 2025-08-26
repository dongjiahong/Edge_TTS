package tts

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"tts-service/internal/config"
	"tts-service/internal/utils"
)

// EdgeTTSClient Edge TTS WebSocket客户端
type EdgeTTSClient struct {
	config *config.EdgeTTSConfig
}

// NewEdgeTTSClient 创建新的Edge TTS客户端
func NewEdgeTTSClient(cfg *config.EdgeTTSConfig) *EdgeTTSClient {
	return &EdgeTTSClient{
		config: cfg,
	}
}

// generateURL 生成动态WebSocket URL
func (c *EdgeTTSClient) generateURL() (string, error) {
	connectionID := strings.ReplaceAll(uuid.New().String(), "-", "")
	trustedClientToken := "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	secMSGECVersion := "1-131.0.2903.99"

	// 生成Sec-MS-GEC
	winEpoch := int64(11644473600) // 从Windows纪元到Unix纪元的偏移量(秒)
	currentTimestamp := time.Now().Unix()
	adjustedSeconds := currentTimestamp + winEpoch
	adjustedSeconds -= adjustedSeconds % 300 // 调整到最近5分钟边界

	// 转换为Windows文件时间(100纳秒单位)
	winFileTime := adjustedSeconds * 10000000

	// 计算SHA-256
	hashInput := fmt.Sprintf("%d%s", winFileTime, trustedClientToken)
	hash := sha256.Sum256([]byte(hashInput))
	secMSGEC := fmt.Sprintf("%X", hash)

	url := fmt.Sprintf("wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1?TrustedClientToken=%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s&ConnectionId=%s",
		trustedClientToken, secMSGEC, secMSGECVersion, connectionID)

	return url, nil
}

// Synthesize 执行语音合成
func (c *EdgeTTSClient) Synthesize(text, voice, format string, speed float64, pitch int) ([]byte, error) {
	// 建立WebSocket连接
	conn, err := c.connect()
	if err != nil {
		return nil, fmt.Errorf("连接Edge TTS失败: %w", err)
	}
	defer conn.Close()

	// 生成请求ID
	requestID := strings.ReplaceAll(uuid.New().String(), "-", "")

	// 发送配置消息
	if err := c.sendConfig(conn, requestID, format); err != nil {
		return nil, fmt.Errorf("发送配置失败: %w", err)
	}

	// 发送SSML文本
	ssml := utils.GenerateSSML(text, voice, speed, pitch)
	if err := c.sendSSML(conn, requestID, ssml); err != nil {
		return nil, fmt.Errorf("发送SSML失败: %w", err)
	}

	// 接收音频数据
	audioData, err := c.receiveAudio(conn, requestID)
	if err != nil {
		return nil, fmt.Errorf("接收音频数据失败: %w", err)
	}

	return audioData, nil
}

// connect 建立WebSocket连接
func (c *EdgeTTSClient) connect() (*websocket.Conn, error) {
	url, err := c.generateURL()
	if err != nil {
		return nil, fmt.Errorf("生成URL失败: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// sendConfig 发送音频配置
func (c *EdgeTTSClient) sendConfig(conn *websocket.Conn, requestID, format string) error {
	// 音频格式映射
	formatMap := map[string]string{
		"mp3": "audio-24khz-48kbitrate-mono-mp3",
		"wav": "riff-24khz-16bit-mono-pcm",
		"ogg": "ogg-24khz-16bit-mono-opus",
	}

	audioFormat, exists := formatMap[format]
	if !exists {
		audioFormat = "audio-24khz-48kbitrate-mono-mp3" // 默认MP3
	}

	config := fmt.Sprintf("X-Timestamp:%s\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n{\"context\":{\"synthesis\":{\"audio\":{\"metadataoptions\":{\"sentenceBoundaryEnabled\":\"false\",\"wordBoundaryEnabled\":\"false\"},\"outputFormat\":\"%s\"}}}}",
		time.Now().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"), audioFormat)

	return conn.WriteMessage(websocket.TextMessage, []byte(config))
}

// sendSSML 发送SSML文本
func (c *EdgeTTSClient) sendSSML(conn *websocket.Conn, requestID, ssml string) error {
	message := fmt.Sprintf("X-Timestamp:%s\r\nX-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nPath:ssml\r\n\r\n%s",
		time.Now().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"), requestID, ssml)

	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// receiveAudio 接收音频数据  
func (c *EdgeTTSClient) receiveAudio(conn *websocket.Conn, requestID string) ([]byte, error) {
	audioChunks := [][]byte{}
	
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		switch messageType {
		case websocket.TextMessage:
			messageStr := string(message)
			if strings.Contains(messageStr, "Path:turn.start") {
				// 开始接收音频
				continue
			} else if strings.Contains(messageStr, "Path:turn.end") {
				// 音频接收完成
				if len(audioChunks) == 0 {
					return nil, fmt.Errorf("未收到音频数据")
				}
				return c.concatenateAudio(audioChunks), nil
			}

		case websocket.BinaryMessage:
			// 提取音频数据 (跳过头部信息)
			audioSeparator := []byte("Path:audio\r\n")
			if idx := c.indexOf(message, audioSeparator); idx >= 0 {
				audioData := message[idx+len(audioSeparator):]
				if len(audioData) > 0 {
					audioChunks = append(audioChunks, audioData)
				}
			}
		}
	}
}

// concatenateAudio 合并音频数据
func (c *EdgeTTSClient) concatenateAudio(chunks [][]byte) []byte {
	totalLength := 0
	for _, chunk := range chunks {
		totalLength += len(chunk)
	}

	result := make([]byte, totalLength)
	offset := 0
	for _, chunk := range chunks {
		copy(result[offset:], chunk)
		offset += len(chunk)
	}

	return result
}

// indexOf 查找字节序列位置
func (c *EdgeTTSClient) indexOf(data, separator []byte) int {
	if len(separator) == 0 {
		return 0
	}

	for i := 0; i <= len(data)-len(separator); i++ {
		if string(data[i:i+len(separator)]) == string(separator) {
			return i
		}
	}
	return -1
}