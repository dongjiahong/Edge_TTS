package tts

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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

// Synthesize 执行语音合成
func (c *EdgeTTSClient) Synthesize(text, voice, format string, speed float64, pitch int) ([]byte, error) {
	// 建立WebSocket连接
	conn, err := c.connect()
	if err != nil {
		return nil, fmt.Errorf("连接Edge TTS失败: %w", err)
	}
	defer conn.Close()

	// 生成请求ID
	requestID := utils.GenerateRequestID()

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
	audioData, err := c.receiveAudio(conn)
	if err != nil {
		return nil, fmt.Errorf("接收音频数据失败: %w", err)
	}

	return audioData, nil
}

// connect 建立WebSocket连接
func (c *EdgeTTSClient) connect() (*websocket.Conn, error) {
	headers := http.Header{}
	headers.Set("User-Agent", c.config.UserAgent)
	headers.Set("Origin", "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold")
	headers.Set("Accept-Encoding", "gzip, deflate, br")
	headers.Set("Accept-Language", "en-US,en;q=0.9")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Pragma", "no-cache")

	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial(c.config.Endpoint, headers)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// sendConfig 发送音频配置
func (c *EdgeTTSClient) sendConfig(conn *websocket.Conn, requestID, format string) error {
	// 音频格式映射
	audioFormat := c.mapAudioFormat(format)
	
	configMsg := fmt.Sprintf(`Path: speech.config
Content-Type: application/json; charset=utf-8
X-RequestId: %s
X-Timestamp: %s

{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":"false","wordBoundaryEnabled":"true"},"outputFormat":"%s"}}}}`,
		requestID,
		c.getTimestamp(),
		audioFormat,
	)

	return conn.WriteMessage(websocket.TextMessage, []byte(configMsg))
}

// sendSSML 发送SSML文本
func (c *EdgeTTSClient) sendSSML(conn *websocket.Conn, requestID, ssml string) error {
	ssmlMsg := fmt.Sprintf(`Path: ssml
Content-Type: application/ssml+xml
X-RequestId: %s
X-Timestamp: %s

%s`, requestID, c.getTimestamp(), ssml)

	return conn.WriteMessage(websocket.TextMessage, []byte(ssmlMsg))
}

// receiveAudio 接收音频数据
func (c *EdgeTTSClient) receiveAudio(conn *websocket.Conn) ([]byte, error) {
	var audioBuffer bytes.Buffer
	
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				break
			}
			return nil, err
		}

		switch messageType {
		case websocket.TextMessage:
			// 处理文本消息（可能包含边界信息等）
			if c.isEndMessage(message) {
				log.Printf("收到结束消息: %s", string(message))
				goto done
			}
		case websocket.BinaryMessage:
			// 处理二进制音频数据
			if c.isAudioData(message) {
				// 跳过头部，提取音频数据
				audioData := c.extractAudioData(message)
				if len(audioData) > 0 {
					audioBuffer.Write(audioData)
				}
			}
		}
	}

done:
	if audioBuffer.Len() == 0 {
		return nil, fmt.Errorf("没有收到音频数据")
	}

	return audioBuffer.Bytes(), nil
}

// mapAudioFormat 映射音频格式
func (c *EdgeTTSClient) mapAudioFormat(format string) string {
	switch strings.ToLower(format) {
	case "mp3":
		return "audio-24khz-48kbitrate-mono-mp3"
	case "wav":
		return "riff-24khz-16bit-mono-pcm"
	case "ogg":
		return "ogg-24khz-16bit-mono-opus"
	default:
		return "audio-24khz-48kbitrate-mono-mp3"
	}
}

// getTimestamp 获取时间戳
func (c *EdgeTTSClient) getTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

// isEndMessage 检查是否为结束消息
func (c *EdgeTTSClient) isEndMessage(message []byte) bool {
	return strings.Contains(string(message), "Path:turn.end")
}

// isAudioData 检查是否为音频数据
func (c *EdgeTTSClient) isAudioData(message []byte) bool {
	// 检查是否包含音频数据头部
	return strings.Contains(string(message[:min(200, len(message))]), "Path:audio") ||
		   bytes.Contains(message[:min(50, len(message))], []byte("Path:audio"))
}

// extractAudioData 提取音频数据
func (c *EdgeTTSClient) extractAudioData(message []byte) []byte {
	// 查找音频数据开始位置（通常在两个换行符之后）
	headerEnd := bytes.Index(message, []byte("\r\n\r\n"))
	if headerEnd == -1 {
		headerEnd = bytes.Index(message, []byte("\n\n"))
		if headerEnd == -1 {
			return nil
		}
		headerEnd += 2
	} else {
		headerEnd += 4
	}
	
	if headerEnd >= len(message) {
		return nil
	}
	
	return message[headerEnd:]
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}