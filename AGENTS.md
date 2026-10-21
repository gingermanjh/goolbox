# AGENTS.md

프로젝트 지침의 단일 출처는 **[CLAUDE.md](CLAUDE.md)** 입니다.

> `CLAUDE.md`는 로컬 전용 파일로 관리됩니다(.gitignore 대상). 이 저장소를 새로
> 클론했다면 `CLAUDE.md`가 없으므로 아래 요약을 기준으로 작업하세요.

## 빠른 요약

- **Go CLI 프로젝트.** 앱은 `cmd/<app>/`, 패키지는 `internal/` 아래에 있습니다.
- **ffmpeg-converter는 macOS 26(Tahoe) + Apple silicon 전용.**
  변환은 Apple `container` 런타임 안의 Linux 컨테이너에서 수행하므로
  호스트에 ffmpeg·Homebrew·Docker가 필요 없습니다. 구 macOS / Intel Mac은 지원 대상이 아닙니다.
- **의존성 자동 관리**: 첫 실행 시 `container` 런타임을 자동 설치하고(관리자 인증 필요),
  변환이 끝나면 컨테이너·microVM·백그라운드 서비스를 정리합니다.
- **커밋 규칙**: conventional commit, 영문 subject(72자 이내), 명령형,
  `<type>/<kebab-case>` 브랜치, `main` 직접 커밋·푸시 금지.
  검사 도구: `prek` (설치: `prek install -t pre-commit -t commit-msg -t pre-push`).
- **품질 게이트**: `just audit` (tidy + vet + lint + test). 커밋 전 훅이
  gofmt / `go mod tidy` / `go vet` / staticcheck / `go test -race` 를 검사합니다.
- **주석과 커밋 메시지는 영어로 작성.**

## 더 자세한 지침

빌드 명령, 코드 스타일, 앱 추가 절차 등은 [CLAUDE.md](CLAUDE.md) 를 참고하세요.
