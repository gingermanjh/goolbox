# ffmpeg-converter

ffmpeg를 사용하여 비디오 파일을 MP4(H.264)로 일괄 변환하는 CLI 도구입니다.

## 특징

- **두 가지 실행 모드**: Docker 컨테이너 또는 로컬 ffmpeg 사용
- **일괄 변환**: 디렉토리 내 모든 비디오 파일을 한 번에 변환
- **프로파일 기반 인코딩**: 미리 정의된 프로파일로 간편한 설정
- **의존성 자동 관리**: Homebrew, ffmpeg, Docker 자동 설치
- **Finder 호환**: 더블클릭으로 실행 가능

## 요구 사항

- macOS (Linux 부분 지원)
- 의존성은 실행 시 자동 설치됨:
  - Docker 모드: Docker Desktop
  - Local 모드: ffmpeg

## 설치

### 소스에서 빌드

```bash
cd cmd/ffmpeg-converter
make build
```

### GOPATH에 설치

```bash
cd cmd/ffmpeg-converter
make install
```

### 크로스 컴파일

```bash
cd cmd/ffmpeg-converter
make build/all    # darwin, linux, windows 전체
```

## 빠른 시작

### 기본 사용법

```bash
# 현재 디렉토리의 .avi 파일을 ./results/에 변환 (standard 프로파일 사용)
./ffmpeg-converter

# 도움말 보기
./ffmpeg-converter -h

# 프로파일 목록 보기
./ffmpeg-converter -list-profiles
```

### 주요 사용 시나리오

```bash
# Docker 대신 로컬 ffmpeg 사용
./ffmpeg-converter -local

# MKV 파일을 오디오 포함하여 변환
./ffmpeg-converter -ext mkv -audio

# 특정 디렉토리 지정
./ffmpeg-converter -input /path/to/videos -output /path/to/output

# 실행 전 명령어 확인 (dry-run)
./ffmpeg-converter -dry-run
```

자세한 사용법은 [USAGE.md](USAGE.md)를 참조하세요.

## 프로파일

인코딩 설정은 프로파일로 관리됩니다.

| 프로파일 | CRF | Preset | Bitrate | Scale | Audio | 설명 |
|----------|-----|--------|---------|-------|-------|------|
| `standard` | 18 | slow | 3000k | 1920:1080 | false | 기본 설정 |

```bash
# 프로파일 목록 확인
./ffmpeg-converter -list-profiles

# 특정 프로파일 사용 (현재는 standard만 제공)
./ffmpeg-converter -profile standard
```

## 기본 동작

| 항목 | 기본값 |
|------|--------|
| 입력 디렉토리 | 실행 파일 위치 |
| 출력 디렉토리 | `./results/` |
| 입력 확장자 | `.avi` |
| 실행 모드 | Docker |
| 프로파일 | `standard` |
| 오디오 | 프로파일 설정 따름 |

## 의존성 자동 관리

프로그램 실행 시 필요한 의존성을 자동으로 확인하고 설치합니다.

### Docker 모드 (기본)

1. Homebrew 확인 (없으면 설치)
2. Docker 확인 (없으면 설치)
3. Docker 데몬 확인 (중지 상태면 자동 시작)

### Local 모드 (`-local`)

1. Homebrew 확인 (없으면 설치)
2. ffmpeg 확인 (없으면 `brew install ffmpeg`)

### 의존성 관련 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-auto-install` | `true` | 의존성 자동 설치 활성화 |
| `-no-auto-install` | `false` | 의존성 체크 완전히 건너뜀 |
| `-y` | `false` | 설치 확인 프롬프트 생략 |

```bash
# 확인 없이 자동 설치
./ffmpeg-converter -y

# 의존성 체크 생략 (이미 설치된 경우)
./ffmpeg-converter -no-auto-install
```

## 프로젝트 구조

```
cmd/ffmpeg-converter/
├── main.go              # CLI 엔트리포인트
├── Makefile             # 빌드 스크립트
├── README.md            # 이 파일
└── USAGE.md             # 상세 사용법

internal/
├── ffmpeg/              # ffmpeg 변환 로직
│   ├── converter.go
│   └── profile.go       # 프로파일 정의
└── deps/                # 의존성 관리
    ├── deps.go          # Manager 및 인터페이스
    ├── platform.go      # 플랫폼 감지
    ├── homebrew.go      # Homebrew 의존성
    ├── ffmpeg.go        # ffmpeg 의존성
    ├── docker.go        # Docker 의존성
    └── installer_darwin.go  # macOS 설치 유틸리티
```

## 관련 문서

- [USAGE.md](USAGE.md) - 전체 옵션 및 상세 사용 예시
