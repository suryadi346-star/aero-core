
# 🌬️ AeroCore AI
> ⚡ Lightweight, Memory-Optimized AI Chat Infrastructure for Low-Resource Environments

![AeroCore AI Logo](assets/logo.png)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/suryadi346-star/aero-core/actions/workflows/ci.yml/badge.svg)](https://github.com/suryadi346-star/aero-core/actions)

**AeroCore AI** adalah fondasi backend AI chat yang dirancang khusus untuk berjalan efisien pada lingkungan dengan **RAM terbatas (3–4 GB)**. Dibangun dengan Go murni, tanpa dependensi berat, dan terintegrasi native dengan **Ollama** & model **Qwen**. Fokus utama: stabilitas, keamanan memori, dan skalabilitas modular.

---

## ✨ Fitur Utama
- 🚀 **Binary Ringan** (~12MB, zero CGO, pure Go SQLite)
-  **RAM-Bounded** (`runtime.SetMemoryLimit`, GC tuning, pool DB ketat)
-  **Ollama Native** (SSE streaming, context window truncation, model switching)
- 🗄️ **SQLite Caching** (Prompt hashing SHA256, response cache, session TTL)
- 🛡️ **Production Hardened** (Rate limiter, panic recovery, SSE sanitization, context cancellation)
- 📊 **Observability** (Prometheus-compatible `/metrics`, deep health checks)
- 🐳 **Docker Ready** (Multi-stage build, memory-limited compose, non-root runtime)
-  **CI/CD Pipeline** (GitHub Actions: vet → build → verify)

---

## 🏗️ Arsitektur
```
[Client / UI]
   ↓ HTTPS/SSE
[Go API Gateway] → Logging → Recovery → RateLimit → Router
   ↓
[Session Manager] → In-Memory LRU + TTL Cleanup + SQLite Persistence
   ↓
[Cache Layer] → Prompt Hashing → Response Cache → Prune (1k entries)
   ↓
[Ollama Client] → Context-Aware Streaming → Graceful Abort on Disconnect
   ↓
[SQLite WAL] → Low-I/O, Safe Concurrency, Busy Timeout 5s
```

---

##  Quick Start

### Prasyarat
- Go 1.22+
- Ollama (v0.4+)
- Docker & Docker Compose (opsional)

### Local Development
```bash
git clone https://github.com/suryadi346-star/aero-core.git
cd aero-core

# Unduh model ringan (disarankan untuk RAM 3-4GB)
ollama pull qwen2.5:1.5b-instruct-q4_k_m

# Jalankan service
go mod tidy
go run ./cmd/server
```

### Docker Deployment
```bash
docker compose up -d
# App akan otomatis terhubung ke Ollama di dalam network
curl http://localhost:8080/health/deep
```

---

## 🔌 API Endpoints

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `GET`  | `/health` | Status dasar service |
| `GET`  | `/health/deep` | Cek konektivitas DB & Ollama |
| `GET`  | `/metrics` | Runtime metrics (Prometheus format) |
| `POST` | `/chat/stream` | SSE streaming chat |

### Contoh Request Chat
```bash
curl -N -X POST http://localhost:8080/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user_001",
    "message": "Jelaskan cara kerja AI ringan di perangkat terbatas?",
    "model": "qwen2.5:1.5b-instruct-q4_k_m"
  }'
```

**Respons (SSE):**
```
event: chunk
data: {"content":"AI ringan bekerja dengan..."}

event: done
data: {}
```

---

## ⚙️ Konfigurasi
File utama: `configs/app.yaml`

| Key | Default | Keterangan |
|-----|---------|------------|
| `system.memory_hard_limit_mb` | `3200` | Hard limit RAM (Go 1.19+) |
| `cache.sqlite_path` | `./data/chat.db` | Path database session & cache |
| `cache.ttl_minutes` | `45` | Sesi aktif sebelum dihapus |
| `model.num_ctx` | `2048` | Context window (disesuaikan RAM) |
| `system.ollama_host` | `127.0.0.1:11434` | Alamat Ollama runtime |

Override via environment variable:
```bash
export CONFIG_PATH=/etc/aero/app.yaml
export OLLAMA_HOST=http://ollama:11434
```

---

## 📊 Performa & Resource
| Komponen | Footprint (Idle) | Maksimal |
|----------|------------------|----------|
| Go Backend | ~15 MB | 800 MB (Docker limit) |
| Ollama + Qwen 1.5B Q4_K_M | ~1.3 GB | 2.5 GB (Docker limit) |
| SQLite + Cache | < 50 MB | ~200 MB (1k entries) |
| **Total Sistem** | **~1.35 GB** | **≤ 3.5 GB** |

Optimasi yang diterapkan:
- `runtime.SetMemoryLimit` & `GOGC=50`
- SQLite `WAL` mode + `busy_timeout=5000`
- Channel drain via `ctx.Done()` (zero goroutine leak)
- Rate limiter capped pada 500 IP aktif
- Context truncation konservatif (`1 token ≈ 2.5 chars`)

---

## 🛠️ Development & CI
Pipeline otomatis pada setiap push/PR:
```yaml
✅ Checkout → Setup Go 1.22 → go mod download
✅ go vet ./... → Static analysis
✅ CGO_ENABLED=0 go build → Static binary
✅ file verification → Pastikan statically linked
```

Jalankan test lokal:
```bash
go vet ./...
go build -ldflags="-s -w" -o aero-core ./cmd/server
./aero-core
```

---

## 📜 Lisensi
Distribusi di bawah [**MIT License**](https://github.com/suryadi346-star/aero-core/blob/5e5d78011d1648bacef2b1d00133ac89db473389/LICENSE) Bebas digunakan, dimodifikasi, dan didistribusikan untuk keperluan komersial maupun open-source. Lihat file `LICENSE` untuk detail.

---

## 🤝 Kontribusi
1. Fork repository
2. Buat branch fitur (`git checkout -b feat/nama-fitur`)
3. Commit perubahan (`git commit -m 'feat: tambah X'`)
4. Push ke branch (`git push origin feat/nama-fitur`)
5. Buka Pull Request

Pastikan `go vet ./...` bersih dan footprint RAM tidak melebihi baseline sebelum merge.

---

**AeroCore AI** → Dibangun untuk efisiensi, dijalankan untuk keandalan.  
 Repository: [github.com/suryadi346-star/aero-core](https://github.com/suryadi346-star/aero-core)  
📩 Contact: [@suryadi346-star](https://github.com/suryadi346-star)


- *aero-core*
