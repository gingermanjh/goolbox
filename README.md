# goolbox

Go로 작성된 CLI 애플리케이션 모음입니다.

## 애플리케이션

| 앱 | 설명 | 문서 |
|----|------|------|
| [ffmpeg-converter](cmd/ffmpeg-converter/) | ffmpeg를 이용한 비디오 일괄 변환 도구 | [README](cmd/ffmpeg-converter/README.md) |

## 요구 사항

- Go 1.25 이상
- macOS 26 (Tahoe) 이상, Apple silicon — `ffmpeg-converter`는 Apple `container` 런타임에 의존합니다.
- [just](https://just.systems) — 태스크 러너 (mise: `mise use -g just`)

## 프로젝트 구조

```
goolbox/
├── cmd/                         # 애플리케이션 진입점
│   └── ffmpeg-converter/        # 비디오 변환 CLI
│       ├── main.go
│       ├── README.md
│       └── USAGE.md
├── internal/                    # 내부 패키지
│   ├── ffmpeg/                  # ffmpeg 변환 로직
│   └── deps/                    # 의존성 자동 관리
├── justfile                     # 빌드/테스트 태스크
└── go.mod
```

## 빌드

```bash
# 사용 가능한 레시피 목록
just

# ffmpeg-converter 빌드
just build

# 빌드 후 CLI 도움말
just run

# $GOPATH/bin에 설치
just install
```

## 개발

```bash
# 테스트 실행
just test

# 코드 품질 검사 (lint, vet, test)
just audit

# 코드 정리 (go mod tidy, go fmt)
just tidy
```

## justfile 레시피

| 레시피 | 설명 |
|--------|------|
| `just` | 레시피 목록 표시 |
| `just test` | 전체 테스트 실행 (race detector) |
| `just test-verbose` | 테스트를 상세 출력으로 실행 |
| `just cover` | 테스트 + 커버리지 리포트 생성 |
| `just lint` | staticcheck 실행 |
| `just vet` | go vet 실행 |
| `just tidy` | go mod tidy + go fmt |
| `just audit` | 전체 품질 검사 (tidy + vet + lint + test) |
| `just build` | ffmpeg-converter 바이너리 빌드 |
| `just run` | 빌드 후 CLI 도움말 출력 |
| `just install` | $GOPATH/bin에 설치 |
| `just version` | 바이너리에 심길 버전 정보 표시 |
| `just clean` | 빌드 결과물 삭제 |

## 커밋 훅

`prek`로 커밋 메시지·브랜치 규칙과 Go 품질 검사를 강제합니다.

```bash
# 훅 설치
prek install -t pre-commit -t commit-msg -t pre-push

# 훅 직접 실행
prek run --all-files
```

규칙은 로컬 `prek.toml`에 있습니다(저장소에 커밋되지 않음).

## 라이선스

MIT
