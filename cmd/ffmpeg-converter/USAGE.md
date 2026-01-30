# ffmpeg-converter 사용법

## 목차

- [기본 실행](#기본-실행)
- [프로파일](#프로파일)
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

인코딩 설정은 프로파일로 관리됩니다. 프로파일은 CRF, 프리셋, 비트레이트, 해상도, 오디오 설정을 포함합니다.

### 프로파일 목록 보기

```bash
./ffmpeg-converter -list-profiles
```

출력 예시:
```
Available encoding profiles:

  standard
    Description: Default encoding settings
    CRF: 18, Preset: slow, Bitrate: 3000k, Scale: 1920:1080, Audio: false
```

### 프로파일 사용

```bash
# 기본 프로파일(standard) 사용
./ffmpeg-converter

# 명시적으로 프로파일 지정
./ffmpeg-converter -profile standard
```

### 내장 프로파일

| 프로파일 | CRF | Preset | Bitrate | Scale | Audio | 설명 |
|----------|-----|--------|---------|-------|-------|------|
| `standard` | 18 | slow | 3000k | 1920:1080 | false | 기본 설정 |

### 오디오 오버라이드

프로파일의 오디오 설정을 오버라이드하려면 `-audio` 플래그를 사용합니다:

```bash
# standard 프로파일 + 오디오 포함
./ffmpeg-converter -audio
```

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
  -audio
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
