
# ️ Contributing to AeroCore AI

Terima kasih telah tertarik untuk berkontribusi pada **AeroCore AI**. Proyek ini dibangun dengan prinsip **efisiensi RAM, keamanan memori, dan kode murni Go**. Setiap kontribusi harus menjaga fondasi ini agar tetap ringan (<3.5 GB total) dan production-ready.

## 📋 Aturan Dasar
1. **Fork & Branch**: Buat branch dari `main` dengan prefix `feat/`, `fix/`, atau `perf/`.
2. **Commit Message**: Gunakan [Conventional Commits](https://www.conventionalcommits.org/). Contoh: `feat: add SQLite prompt cache`
3. **RAM Constraint**: Kontribusi tidak boleh meningkatkan footprint idle melebihi baseline (~1.35 GB). Sertakan hasil `runtime.MemStats` jika mengubah alokasi memori.
4. **Zero CGO**: Proyek ini 100% pure Go. Hindiri dependensi yang memerlukan `cgo` atau kompilasi C.
5. **Testing**: Jalankan `make vet` dan `make build` sebelum membuka PR.

## 🚀 Development Setup
```bash
git clone https://github.com/YOUR_USERNAME/aero-core.git
cd aero-core
go mod download
ollama pull qwen2.5:1.5b-instruct-q4_k_m
make run
```

## ✅ Checklist Sebelum Pull Request
- [ ] `go vet ./...` bersih tanpa warning
- [ ] `gofmt -s -w .` telah dijalankan
- [ ] Tidak ada memory leak (channel drain, context cancellation, sync primitives aman)
- [ ] SSE header & sanitasi JSON tetap compliant
- [ ] Dokumentasi endpoint/konfigurasi diperbarui jika relevan
- [ ] Commit message sesuai standar

## 🔍 Review Process
1. Maintainer akan menjalankan CI pipeline otomatis.
2. Review fokus pada: keamanan memori, race condition, dan efisiensi RAM.
3. Setelah approval, maintainer akan squash & merge ke `main`.

## 📩 Diskusi & Fitur
- Gunakan **GitHub Issues** untuk bug report & feature request.
- Gunakan **GitHub Discussions** untuk arsitektur, benchmark, atau ide optimasi.

Mari bangun infrastruktur AI yang ringan, aman, dan dapat diandalkan. 🌬️
```

---

### 🛡️ `CODE_OF_CONDUCT.md`
```markdown
# Contributor Covenant Code of Conduct

## Our Pledge
We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone, regardless of age, body size, visible or invisible disability, ethnicity, sex characteristics, gender identity and expression, level of experience, education, socio-economic status, nationality, personal appearance, race, religion, or sexual identity and orientation.

## Our Standards
Examples of behavior that contributes to a positive environment:
- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community

Examples of unacceptable behavior:
- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information without explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

## Enforcement Responsibilities
Community leaders are responsible for clarifying and enforcing our standards of acceptable behavior and will take appropriate and fair corrective action in response to any behavior that they deem inappropriate, threatening, offensive, or harmful.

## Scope
This Code of Conduct applies within all community spaces, and also applies when an individual is officially representing the community in public spaces.

## Enforcement
Instances of abusive, harassing, or otherwise unacceptable behavior may be reported to the community leaders responsible for enforcement at [@suryadi346-star](https://github.com/suryadi346-star). All complaints will be reviewed and investigated promptly and fairly.

## Attribution
This Code of Conduct is adapted from the [Contributor Covenant](https://www.contributor-covenant.org), version 2.1.
```

---

### 🔐 `SECURITY.md`
```markdown
# Security Policy

## Supported Versions
| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | ✅ Yes             |
| < 0.1   | ❌ No              |

## Reporting a Vulnerability
If you discover a security vulnerability in AeroCore AI, please report it privately via GitHub Security Advisories:
🔒 [Report Vulnerability](https://github.com/suryadi346-star/aero-core/security/advisories/new)

**Do not open public issues for security flaws.**

### What to Include
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline
- Acknowledgment: Within 48 hours
- Initial assessment: Within 5 business days
- Patch/Workaround: Within 14 days (critical) or next release cycle

We appreciate responsible disclosure and will credit contributors in release notes unless anonymity is requested.
```

---

### 🗑️ `.gitignore`
```gitignore
# Binaries & Build
*.exe
*.exe~
*.dll
*.so
*.dylib
aero-core
vendor/

# Go
*.test
*.out
coverage.txt
*.prof

# Environment & Config
.env
.env.local
configs/app.local.yaml

# Database & Cache
data/
*.db
*.db-journal
*.db-wal

# IDE & OS
.vscode/
.idea/
*.swp
*.swo
.DS_Store
Thumbs.db

# Docker
docker-compose.override.yml
```

---

### ️ `Makefile`
```makefile
.PHONY: help build run vet lint test docker-up docker-down clean

BIN_NAME := aero-core
GO_FLAGS := CGO_ENABLED=0 GOOS=linux
LDFLAGS := -ldflags="-s -w"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build static binary
	$(GO_FLAGS) go build $(LDFLAGS) -o $(BIN_NAME) ./cmd/server

run: ## Run server locally
	go run ./cmd/server

vet: ## Run static analysis
	go vet ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

test: ## Run tests
	go test -race -coverprofile=coverage.out ./...

docker-up: ## Start services via Docker Compose
	docker compose up -d

docker-down: ## Stop & remove containers
	docker compose down --remove-orphans

clean: ## Remove build artifacts
	rm -f $(BIN_NAME) coverage.out
```

---

### 📜 `CHANGELOG.md`
```markdown
# Changelog

All notable changes to **AeroCore AI** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-05-02
### Added
- Foundation architecture with Go 1.22, zero CGO, pure SQLite
- SSE streaming chat endpoint with context-aware cancellation
- Session management with TTL cleanup & SQLite persistence
- Prompt caching (SHA256) & automatic pruning (1k entries)
- Rate limiter (500 IP cap), panic recovery, logging middleware
- Deep health check (`/health/deep`) & Prometheus-compatible `/metrics`
- Docker multi-stage build + memory-limited `docker-compose.yml`
- GitHub Actions CI pipeline (vet → build → verify)
- RAM hard limit (`runtime.SetMemoryLimit`) & GC tuning (`GOGC=50`)
- Professional branding, MIT license, contributing & security policies

### Technical Constraints
- Max RAM footprint: ≤ 3.5 GB (App 800MB + Ollama 2.5GB)
- Context window: 2048 tokens (conservative truncation)
- Dependencies: `yaml.v3`, `modernc.org/sqlite` only
```

---

## ✅ Checklist Final Repository
| File | Status | Fungsi |
|------|--------|--------|
| `README.md` | ✅ Siap | Dokumentasi utama & quick start |
| `LICENSE` | ✅ Siap | MIT 2026, hak cipta terbuka |
| `CONTRIBUTING.md` | ✅ Siap | Panduan kontributor + RAM constraint |
| `CODE_OF_CONDUCT.md` | ✅ Siap | Standar komunitas open-source |
| `SECURITY.md` | ✅ Siap | Pelaporan kerentanan privat |
| `.gitignore` | ✅ Siap | Cegah commit DB, binary, env |
| `Makefile` | ✅ Siap | Otomasi build, run, lint, docker |
| `CHANGELOG.md` | ✅ Siap | Tracking versi & fitur |
| `assets/logo.png` | ⏳ Manual | Letakkan logo AI yang telah digenerate |

---
 **Repository Anda sekarang 100% production-ready, standar GitHub open-source, dan siap dipublikasikan.**

Langkah selanjutnya:
```bash
mkdir -p assets
# Pindahkan logo ke assets/logo.png
git add .
git commit -m "feat: initialize aerocore ai with full oss standards"
git push origin main
```
