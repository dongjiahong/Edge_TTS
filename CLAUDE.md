# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a TTS (Text-to-Speech) service built in Go that provides both native REST APIs and OpenAI-compatible endpoints. It uses Microsoft Edge TTS as the underlying speech synthesis engine and implements a dual-caching system (SQLite + Redis) for performance optimization.

## Core Architecture

The service follows a layered architecture:

```
HTTP API Layer (Gin) → Service Layer → TTS Engine Layer (Edge TTS WebSocket)
                    ↓
               Caching Layer (Redis + SQLite)
                    ↓
               Database Layer (SQLite)
```

### Key Components

- **WebSocket Client**: `internal/tts/edge_tts.go` implements the Edge TTS WebSocket protocol with dynamic URL generation and Sec-MS-GEC authentication
- **Dual Caching**: `internal/cache/redis.go` and SQLite-based caching to avoid repeated synthesis requests
- **API Compatibility**: `internal/server/openai.go` provides OpenAI TTS API compatibility with voice mapping
- **User Management**: Command-line tool in `cmd/user-manager/` for API key management

### Edge TTS Integration

The Edge TTS client generates dynamic URLs with cryptographic authentication:
- Uses SHA-256 hash of Windows file time + trusted client token
- Implements proper WebSocket message format for SSML synthesis
- Handles binary audio data extraction from WebSocket responses

## Development Commands

### Project Setup
```bash
# Initialize project (creates directories, compiles, generates API keys)
./scripts/init.sh

# Install dependencies and build
go mod tidy
go build -o tts-service main.go
```

### Development Workflow
```bash
# Start service (port 2828)
./scripts/start.sh

# Stop service  
./scripts/stop.sh

# Test all APIs
./scripts/test-api.sh

# User management
./scripts/manage-user.sh create "username"
./scripts/manage-user.sh list
./scripts/manage-user.sh delete "api_key"
```

### Manual Testing
```bash
# Health check
curl http://localhost:2828/api/v1/health

# Native TTS API
curl -X POST http://localhost:2828/api/v1/tts/synthesize \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello world", "voice": "zh-CN-XiaoxiaoNeural", "format": "mp3"}'

# OpenAI-compatible API  
curl -X POST http://localhost:2828/api/v1/audio/speech \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "tts-1", "input": "Hello", "voice": "alloy"}' \
  --output speech.mp3
```

## Configuration

The service is configured via `config.yaml`:
- **Server**: Port (2828), host binding
- **Database**: SQLite path (./data/tts.db)  
- **Storage**: Audio file storage directory and cleanup policies
- **Edge TTS**: WebSocket endpoint and user agent
- **Redis**: Optional caching layer configuration

## Important Implementation Details

### WebSocket Authentication
Edge TTS requires dynamic URL generation with Sec-MS-GEC header based on current time and trusted client token. The implementation in `generateURL()` must maintain compatibility with Microsoft's authentication scheme.

### Voice Mapping
OpenAI voice names (alloy, echo, etc.) are mapped to Edge TTS voices in `internal/server/openai.go`. Adding new voices requires updating both the voice list and mapping functions.

### Caching Strategy
- **L1 Cache**: Redis (1 hour TTL for hot data)
- **L2 Cache**: SQLite (persistent cache with text hash keys)
- Automatic fallback to SQLite-only if Redis unavailable

### Database Schema
Users table stores API keys and metadata. Cache table stores text hashes mapped to audio file paths. Both use SQLite with WAL mode for concurrent access.

## Key Files

- `main.go`: Application entry point and configuration loading
- `internal/tts/edge_tts.go`: Core Edge TTS WebSocket client implementation  
- `internal/server/openai.go`: OpenAI API compatibility layer
- `internal/db/`: Database operations and models
- `scripts/`: Development and deployment automation