# ffmpeg-converter 사용법

## 목차

- [기본 실행](#기본-실행)
- [프로파일](#프로파일)
- [재생 속도 조절](#재생-속도-조절)
- [전체 옵션](#전체-옵션)
- [사용 예시](#사용-예시)
- [의존성 자동 설치](#의존성-자동-설치)
- [문제 해결](#문제-해결)

---

## 기본 실행

```bash
./ffmpeg-converter
```

실행 파일 위치의 `.avi` 파일을 찾아 `./results/` 디렉토리에 MP4로 변환합니다.
기본 프로파일(`standard`)의 인코딩 설정이 적용됩니다.

---

## 프로파일

인코딩 설정은 프로파일로 관리됩니다. 프로파일은 CRF, 프리셋, 비트레이트, 해상도, 오디오, 재생 속도 설정을 포함합니다.

### 프로파일 목록 보기

```bash
./ffmpeg-converter -list-profiles
```

출력 예시:
```
Available encoding profiles:

  2x
    Description: Play video at 2x speed (audio pitch preserved when enabled)
    CRF: 18, Preset: slow, Bitrate: 3000k, Scale: 1920:1080, Audio: false, Speed: 2x

  standard
    Description: Default encoding settings
    CRF: 18, Preset: slow, Bitrate: 3000k, Scale: 1920:1080, Audio: false, Speed: 1x
```

### 프로파일 사용

```bash
# 기본 프로파일(standard) 사용
./ffmpeg-converter

# 명시적으로 프로파일 지정
./ffmpeg-converter -profile standard

# 2배속 프로파일 사용
./ffmpeg-converter -profile 2x
```

### 내장 프로파일

| 프로파일 | CRF | Preset | Bitrate | Scale | Audio | Speed | 설명 |
|----------|-----|--------|---------|-------|-------|-------|------|
| `standard` | 18 | slow | 3000k | 1920:1080 | false | 1x | 기본 설정 |
| `2x` | 18 | slow | 3000k | 1920:1080 | false | 2x | 영상을 2배속으로 재생 |

### 오디오 오버라이드

프로파일의 오디오 설정을 오버라이드하려면 `-audio` 플래그를 사용합니다:

```bash
# standard 프로파일 + 오디오 포함
./ffmpeg-converter -audio
```

---

## 재생 속도 조절

영상을 임의 배속(예: 2배속, 1.5배속, 슬로우모션)으로 변환할 수 있습니다. 두 가지 방법이 있습니다.

### 1. 프로파일 사용

```bash
# 2배속 프로파일
./ffmpeg-converter -profile 2x
```

### 2. `-speed` 플래그로 임의 배속 지정

`-speed`는 프로파일의 Speed 값을 오버라이드합니다.

```bash
# 2배속
./ffmpeg-converter -speed 2.0

# 1.5배속
./ffmpeg-converter -speed 1.5

# 3배속 + 오디오 포함 (피치 유지)
./ffmpeg-converter -speed 3.0 -audio

# 절반 속도 (슬로우모션)
./ffmpeg-converter -speed 0.5
```

### 동작 원리

- **영상**: `setpts=(1/speed)*PTS` 필터로 프레임 타임스탬프를 재계산합니다.
- **오디오** (`-audio` 사용 시): `atempo=speed` 필터로 속도만 변경하고 피치는 유지합니다.
- `atempo`의 유효 범위는 `[0.5, 2.0]`이지만, 그 밖의 값(예: 5배속, 0.25배속)도 자동으로 필터를 체이닝하여 처리합니다.
  - 예: `-speed 5 -audio` → `atempo=2.0,atempo=2.0,atempo=1.25`
  - 예: `-speed 0.25 -audio` → `atempo=0.5,atempo=0.5`

### 동작 확인 (dry-run)

```bash
./ffmpeg-converter -profile 2x -dry-run
./ffmpeg-converter -speed 1.5 -audio -dry-run
```

생성되는 ffmpeg 명령어에서 `-vf setpts=...,scale=...` 및 (오디오 포함 시) `-af atempo=...`가 포함되는지 확인하면 됩니다.

---

## 전체 옵션

### 입출력 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-input` | 실행 파일 위치 | 입력 디렉토리 경로 |
| `-output` | `./results` | 출력 디렉토리 경로 |
| `-ext` | `avi` | 변환할 파일 확장자 (점 제외) |

### 실행 모드 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-docker` | `true` | Docker 컨테이너의 ffmpeg 사용 |
| `-local` | `false` | 로컬에 설치된 ffmpeg 사용 |
| `-dry-run` | `false` | 실제 실행 없이 명령어만 출력 |

### 프로파일 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-profile` | `standard` | 사용할 인코딩 프로파일 |
| `-list-profiles` | `false` | 사용 가능한 프로파일 목록 출력 |
| `-audio` | `false` | 오디오 포함 (프로파일 기본값 오버라이드) |
| `-speed` | `0` | 재생 속도 배율 (예: `2.0`=2배속, `0.5`=절반). `0`이면 프로파일 값 사용 |

### 의존성 관리 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-auto-install` | `true` | 누락된 의존성 자동 설치 |
| `-no-auto-install` | `false` | 의존성 체크 완전히 건너뜀 |
| `-y` | `false` | 설치 확인 프롬프트 생략 |

### 기타 옵션

| 옵션 | 설명 |
|------|------|
| `-version` | 버전 정보 출력 |
| `-h`, `-help` | 도움말 출력 |

---

## 사용 예시

### 기본 변환

```bash
# Docker 모드 (기본)
./ffmpeg-converter

# 로컬 ffmpeg 사용
./ffmpeg-converter -local
```

### 입출력 디렉토리 지정

```bash
# 특정 디렉토리의 파일 변환
./ffmpeg-converter -input /path/to/videos -output /path/to/output

# MKV 파일 변환
./ffmpeg-converter -ext mkv
```

### 프로파일 사용

```bash
# 기본 프로파일
./ffmpeg-converter

# 명시적 프로파일 지정
./ffmpeg-converter -profile standard

# 프로파일 + 오디오 포함
./ffmpeg-converter -profile standard -audio
```

### 재생 속도 조절

```bash
# 2배속 프로파일
./ffmpeg-converter -profile 2x

# 1.5배속 (임의 배속)
./ffmpeg-converter -speed 1.5

# 2배속 + 오디오 포함 (피치 유지)
./ffmpeg-converter -speed 2.0 -audio

# 슬로우모션 (절반 속도)
./ffmpeg-converter -speed 0.5
```

### 테스트 및 확인

```bash
# 프로파일 목록 확인
./ffmpeg-converter -list-profiles

# 실행 전 명령어 확인 (dry-run)
./ffmpeg-converter -dry-run

# 버전 확인
./ffmpeg-converter -version
```

### 의존성 관리

```bash
# 확인 프롬프트 없이 자동 설치
./ffmpeg-converter -y

# 로컬 모드 + 자동 설치 (프롬프트 없음)
./ffmpeg-converter -local -y

# 의존성 체크 건너뛰기 (이미 설치된 경우)
./ffmpeg-converter -no-auto-install
```

### 복합 예시

```bash
# 전체 설정 예시
./ffmpeg-converter \
  -input ./raw_videos \
  -output ./converted \
  -ext mkv \
  -local \
  -profile standard \
  -audio \
  -speed 2.0
```

---

## 의존성 자동 설치

프로그램 실행 시 필요한 의존성을 자동으로 확인하고 설치합니다.

### 시나리오별 동작

#### Docker 모드에서 Docker가 없는 경우

```
✓ Homebrew is installed
✗ Docker is not installed
Install Docker? [y/N]: y
Installing Docker...
✓ Docker installed successfully
Starting Docker...
Waiting for Docker daemon to be ready...
✓ Docker started successfully
```

#### Local 모드에서 ffmpeg이 없는 경우

```
✓ Homebrew is installed
✗ ffmpeg is not installed
Install ffmpeg? [y/N]: y
Installing ffmpeg...
✓ ffmpeg installed successfully
```

#### Docker가 설치되어 있지만 실행 중이 아닌 경우

```
✓ Homebrew is installed
⚠ Docker is installed but stopped
Start Docker? [y/N]: y
Starting Docker...
Waiting for Docker daemon to be ready...
✓ Docker started successfully
```

---

## 문제 해결

### Docker 관련

**문제**: Docker daemon is not running
```bash
# 해결: Docker Desktop을 시작하거나 -y 옵션으로 자동 시작
./ffmpeg-converter -y
```

**문제**: Docker 이미지를 찾을 수 없음
```bash
# 해결: 첫 실행 시 이미지를 자동으로 다운로드합니다
# 네트워크 연결을 확인하세요
```

### ffmpeg 관련

**문제**: ffmpeg: command not found
```bash
# 해결 1: 자동 설치 사용
./ffmpeg-converter -local -y

# 해결 2: 수동 설치
brew install ffmpeg
```

### 파일 관련

**문제**: No .avi files found
```bash
# 해결: 올바른 확장자 지정
./ffmpeg-converter -ext mkv

# 또는 입력 디렉토리 확인
./ffmpeg-converter -input /correct/path
```

**문제**: Permission denied
```bash
# 해결: 출력 디렉토리 권한 확인
chmod 755 ./results

# 또는 다른 출력 디렉토리 지정
./ffmpeg-converter -output ~/Videos/output
```

### 프로파일 관련

**문제**: Unknown profile: xxx
```bash
# 해결: 사용 가능한 프로파일 확인
./ffmpeg-converter -list-profiles
```

### 기타

**문제**: 변환이 너무 느림
```bash
# 해결: 로컬 ffmpeg 사용 (Docker 오버헤드 제거)
./ffmpeg-converter -local
```
