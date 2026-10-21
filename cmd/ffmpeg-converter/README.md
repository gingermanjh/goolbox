# ffmpeg-converter

ffmpeg를 사용하여 비디오 파일을 MP4(H.264)로 일괄 변환하는 CLI 도구입니다.
변환은 Apple의 `container` 런타임 안의 Linux 컨테이너에서 수행하므로, 호스트에 ffmpeg를 설치할 필요가 없습니다.

## 특징

- **컨테이너 기반 변환**: Apple `container` 런타임에서 ffmpeg 이미지(linuxserver/ffmpeg) 실행
- **일괄 변환**: 디렉토리 내 모든 비디오 파일을 한 번에 변환
- **프로파일 기반 인코딩**: 미리 정의된 프로파일로 간편한 설정
- **재생 속도 조절**: 영상을 임의 배속(예: 2배속, 1.5배속)으로 변환. 오디오 포함 시 피치 유지
- **의존성 자동 관리**: Apple `container` 런타임 자동 설치
- **Finder 호환**: 더블클릭으로 실행 가능

## 요구 사항

- **macOS 26 (Tahoe) 이상, Apple silicon** — Apple `container` 런타임의 지원 범위
- 그 외 macOS 버전과 Intel Mac은 지원하지 않습니다.

## 설치

### 소스에서 빌드

```bash
just build
```

### GOPATH에 설치

```bash
just install
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
# MKV 파일을 오디오 포함하여 변환
./ffmpeg-converter -ext mkv -audio

# 특정 디렉토리 지정
./ffmpeg-converter -input /path/to/videos -output /path/to/output

# 2배속 영상으로 변환 (프로파일 사용)
./ffmpeg-converter -profile 2x

# 임의 배속 (1.5배속, 오디오 포함)
./ffmpeg-converter -speed 1.5 -audio

# 실행 전 명령어 확인 (dry-run)
./ffmpeg-converter -dry-run
```

자세한 사용법은 [USAGE.md](USAGE.md)를 참조하세요.

## 프로파일

인코딩 설정은 프로파일로 관리됩니다.

| 프로파일 | CRF | Preset | Scale | Audio | Speed | 설명 |
|----------|-----|--------|-------|-------|-------|------|
| `standard` | 18 | slow | 1920:1080 | false | 1x | 기본 설정 |
| `2x` | 18 | slow | 1920:1080 | false | 2x | 영상을 2배속으로 재생 |

```bash
# 프로파일 목록 확인
./ffmpeg-converter -list-profiles

# 특정 프로파일 사용
./ffmpeg-converter -profile standard
./ffmpeg-converter -profile 2x
```

## 재생 속도 조절

`-speed` 플래그로 임의 배속을 지정할 수 있습니다. 프로파일의 기본 Speed 값을 오버라이드합니다.

```bash
# 2배속
./ffmpeg-converter -speed 2.0

# 1.5배속 + 오디오 포함 (피치 유지)
./ffmpeg-converter -speed 1.5 -audio

# 절반 속도(슬로우모션)
./ffmpeg-converter -speed 0.5
```

- 값이 `0`이거나 지정하지 않으면 프로파일의 Speed 값이 사용됩니다.
- 오디오가 포함된 경우 `atempo` 필터가 적용되어 속도만 바뀌고 피치는 유지됩니다.
- 극단적인 배속(예: 5x, 0.25x)도 자동으로 `atempo` 필터 체이닝으로 처리됩니다.

## 기본 동작

| 항목 | 기본값 |
|------|--------|
| 입력 디렉토리 | 실행 파일 위치 |
| 출력 디렉토리 | `./results/` |
| 입력 확장자 | `.avi` |
| 실행 모드 | Apple container |
| 프로파일 | `standard` |
| 오디오 | 프로파일 설정 따름 |

## 의존성 자동 관리

프로그램 실행 시 Apple `container` 런타임을 자동으로 확인하고 설치합니다.

1. `container` 명령 확인 (없으면 GitHub 릴리스의 공식 설치 패키지를 내려받아 설치)
2. `container` 서비스 상태 확인 (중지 상태면 자동 시작)

설치 패키지는 `/usr/local`에 설치되므로 관리자 인증이 필요합니다.

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

## 설치는 1회, 실행마다 약 1초

`container`는 설치되면 계속 재사용됩니다. 반복해서 설치되거나 다시 다운로드되지 않습니다.

| 작업 | 시점 | 비용 |
|------|------|------|
| `container` 설치 | 첫 실행 시 1회 (없을 때만) | pkg 149MB 다운로드 + 관리자 인증 |
| Linux 커널 다운로드 | 첫 실행 시 1회 | 29MB, 영구 보관 |
| ffmpeg 이미지 pull | 첫 실행 시 1회 | ~950MB, 캐시 유지 |
| 서비스 기동 | 매 실행 | 0.5초 |
| 컨테이너 부팅 | 매 실행 | 0.6초 |

즉 첫 실행 이후에는 **변환 시작까지 약 1.1초**만 소요되고, 영상 개수와 무관합니다.

### 디스크 점유

```
~/Library/Application Support/com.apple.container/
├── content     245MB   이미지 콘텐츠
├── snapshots 693MB   컨테이너 스냅샷
└── kernels      29MB   Linux 커널
```

현재 사용 중인 `latest` 이미지는 캐시에 남습니다. 완전히 비우려면 아래 "변환 후 자원 정리"를 참고하세요.

### 업데이트

자동 업데이트는 하지 않습니다. 설치된 시점의 최신 릴리스로 고정되며, 업데이트가 필요하면 공식 스크립트를 실행하세요.

```bash
container system stop
/usr/local/bin/update-container.sh
container system start
```

## 변환 후 자원 정리

변환이 끝나면 컨테이너 런타임이 남기는 자원을 자동으로 정리합니다.

| 대상 | 처리 |
|------|------|
| 컨테이너 | `--rm`으로 실행 직후 자동 삭제 |
| microVM | 컨테이너 종료와 함께 내려감 |
| 낡은 이미지 | 무태그(dangling) 이미지만 정리 — 현재 쓰는 `latest`는 캐시로 유지 |
| 백그라운드 서비스 | `container-apiserver` 등 종료 (약 45MB frees) |

`-keep-services`를 주면 서비스를 종료하지 않고 남겨둡니다. 반복 실행할 때
서비스 기동(0.5초)을 아끼려면 사용하세요.

```bash
# 서비스는 남겨두고 변환만 수행
./ffmpeg-converter -keep-services
```

디스크를 완전히 비우려면(현재 이미지까지 삭제):

```bash
container system start
container image prune --all
container system stop
```

## 프로젝트 구조

```
cmd/ffmpeg-converter/
├── main.go              # CLI 엔트리포인트
├── README.md            # 이 파일
└── USAGE.md             # 상세 사용법

internal/
├── ffmpeg/              # ffmpeg 변환 로직
│   ├── converter.go
│   └── profile.go       # 프로파일 정의
└── deps/                # 의존성 관리
    ├── deps.go          # Manager 및 인터페이스
    ├── platform.go      # 플랫폼 감지
    └── container.go     # Apple container 의존성
```

## 관련 문서

- [USAGE.md](USAGE.md) - 전체 옵션 및 상세 사용 예시
