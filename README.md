# goolbox

Go로 작성된 CLI 애플리케이션 모음입니다.

## 애플리케이션

| 앱 | 설명 | 문서 |
|----|------|------|
| [ffmpeg-converter](cmd/ffmpeg-converter/) | ffmpeg를 이용한 비디오 일괄 변환 도구 | [README](cmd/ffmpeg-converter/README.md) |

## 요구 사항

- Go 1.21 이상
- macOS 또는 Linux

## 프로젝트 구조

```
goolbox/
├── cmd/                         # 애플리케이션 진입점
│   └── ffmpeg-converter/        # 비디오 변환 CLI
│       ├── main.go
│       ├── Makefile             # 앱별 빌드
│       ├── README.md
│       └── USAGE.md
├── internal/                    # 내부 패키지
│   ├── ffmpeg/                  # ffmpeg 변환 로직
│   └── deps/                    # 의존성 자동 관리
├── Makefile                     # 프로젝트 전체 빌드/테스트
└── go.mod
```

## 빌드

### 전체 앱 빌드

```bash
# 현재 플랫폼용
make build

# 모든 플랫폼용 (darwin, linux, windows)
make build/all
```

### 특정 앱 빌드

```bash
# 방법 1: 루트에서
make -C cmd/ffmpeg-converter build

# 방법 2: 앱 디렉토리에서
cd cmd/ffmpeg-converter
make build
```

## 개발

```bash
# 테스트 실행
make test

# 코드 품질 검사 (lint, vet, test)
make audit

# 코드 정리 (go mod tidy, go fmt)
make tidy
```

## Makefile 타겟

### 루트 Makefile

| 타겟 | 설명 |
|------|------|
| `make test` | 전체 테스트 실행 |
| `make lint` | staticcheck 실행 |
| `make vet` | go vet 실행 |
| `make tidy` | go mod tidy + go fmt |
| `make audit` | 전체 품질 검사 |
| `make build` | 모든 앱 빌드 (현재 플랫폼) |
| `make build/all` | 모든 앱 빌드 (모든 플랫폼) |
| `make clean` | 빌드 결과물 삭제 |
| `make apps` | 앱 목록 표시 |

### 앱별 Makefile

각 앱 디렉토리(`cmd/<app>/`)에서 사용:

| 타겟 | 설명 |
|------|------|
| `make build` | 현재 플랫폼용 빌드 |
| `make run` | 빌드 후 실행 |
| `make install` | $GOPATH/bin에 설치 |
| `make build/all` | 모든 플랫폼용 빌드 |
| `make clean` | 빌드 결과물 삭제 |
| `make version` | 버전 정보 표시 |

## 라이선스

MIT
