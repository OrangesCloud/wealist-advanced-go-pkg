# Wealist Go Common Library

Go 서비스들을 위한 공통 라이브러리입니다.

## 모듈 구성

| 패키지 | 설명 |
|--------|------|
| `config` | YAML + 환경변수 기반 설정 로더 |
| `middleware` | Gin 미들웨어 (Logger, Recovery, Metrics, CORS) |
| `response` | 표준 API 응답 포맷 |
| `health` | K8s 헬스체크 핸들러 (/health, /ready) |
| `logger` | Zap 로거 설정 |

---

## 사용 방법

### 1. 모노레포 환경 (현재)

`go.work` 파일이 루트에 있으면 자동으로 로컬 모듈 참조:

```go
import "github.com/wealist/common/config"
```

### 2. 멀티레포 환경 (분리 후)

서비스를 별도 레포로 분리한 후에는 아래 단계를 따르세요.

#### Step 1: 공통 라이브러리 레포 생성

```bash
# 새 레포 생성 후 pkg/common 내용 복사
git clone https://github.com/YOUR_ORG/wealist-go-common.git
cp -r pkg/common/* wealist-go-common/
cd wealist-go-common

# go.mod의 모듈명 변경
sed -i 's|github.com/wealist/common|github.com/YOUR_ORG/wealist-go-common|g' go.mod

# 태그 생성 (버전 관리)
git add . && git commit -m "Initial commit"
git tag v1.0.0
git push origin main --tags
```

#### Step 2: 서비스에서 의존성 추가

```bash
cd your-service

# go.mod에서 replace 제거하고 실제 의존성 추가
go get github.com/YOUR_ORG/wealist-go-common@v1.0.0

# import 경로 변경
find . -name "*.go" -exec sed -i 's|github.com/wealist/common|github.com/YOUR_ORG/wealist-go-common|g' {} \;
```

#### Step 3: Private 레포인 경우

```bash
# GOPRIVATE 환경변수 설정
export GOPRIVATE=github.com/YOUR_ORG/*

# 또는 ~/.gitconfig에 추가
git config --global url."git@github.com:YOUR_ORG/".insteadOf "https://github.com/YOUR_ORG/"
```

---

## 마이그레이션 체크리스트

서비스를 멀티레포로 분리할 때:

- [ ] 공통 라이브러리 별도 레포 생성
- [ ] go.mod 모듈명 변경
- [ ] 버전 태그 생성 (v1.0.0)
- [ ] 서비스 레포에서 go.work 삭제
- [ ] 서비스 go.mod에서 replace 제거
- [ ] import 경로 업데이트
- [ ] CI/CD에서 GOPRIVATE 설정 (private 레포인 경우)

---

## 로컬 개발 팁

### 공통 라이브러리 수정하면서 서비스 테스트

멀티레포 환경에서도 로컬 개발 시 go.work 또는 replace 사용 가능:

```bash
# 방법 1: go.work 사용 (권장)
# 작업 디렉토리에 go.work 생성
go work init
go work use ./wealist-go-common
go work use ./board-service

# 방법 2: replace 사용
# board-service/go.mod에 추가
# replace github.com/YOUR_ORG/wealist-go-common => ../wealist-go-common
```

---

## 버전 관리

Semantic Versioning을 따릅니다:

- **MAJOR**: 호환성 깨지는 변경 (v2.0.0)
- **MINOR**: 새 기능 추가 (v1.1.0)
- **PATCH**: 버그 수정 (v1.0.1)

```bash
# 새 버전 릴리스
git tag v1.1.0
git push origin v1.1.0

# 서비스에서 업데이트
go get github.com/YOUR_ORG/wealist-go-common@v1.1.0
```
