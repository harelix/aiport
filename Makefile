.PHONY: build web-build test check release check-offline dev leaks hooks docs docs-dev

VERSION ?= 0.0.0-dev
# Pure-Go build: no C toolchain or system libraries needed, on any platform.
export CGO_ENABLED = 0
# Never download during builds: everything comes from vendor/.
export GOFLAGS = -mod=vendor
# UPSTREAM (owner/name on GitHub) is where "Suggest to aiport" and "Share
# upstream" drafts open unless aiport.yaml sets upstream.repository.
UPSTREAM ?=
LDFLAGS = -s -w -X main.version=$(VERSION) -X github.com/harelix/aiport/internal/config.BuiltInUpstream=$(UPSTREAM)
PLATFORMS = darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/amd64

# Live development: Go rebuild/restart on save + Vite dashboard with live reload.
dev:
	./scripts/dev.sh

web-build:
	pnpm --dir web build

# Documentation site (site/, Astro Starlight), built from docs/ into site/dist.
# Needs `pnpm --dir site install` once. CI publishes it to GitHub Pages from main.
docs:
	pnpm --dir site build

docs-dev:
	pnpm --dir site dev

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/aiport ./cmd/aiport

test:
	go test ./...

check:
	@unformatted=$$(gofmt -l cmd internal); if [ -n "$$unformatted" ]; then echo "gofmt -w needed:"; echo "$$unformatted"; exit 1; fi
	go vet ./...

# Proves a fresh clone builds with no network: vendored Go deps + committed UI.
check-offline:
	GOPROXY=off go build -trimpath -o /dev/null ./cmd/aiport
	@test -f internal/dashboard/dist/index.html || (echo "dashboard dist missing: run make web-build and commit it" && exit 1)

# Self-contained downloads: one executable per platform (UI embedded), archived
# with README and licenses, plus SHA-256 checksums.
release: check-offline
	rm -rf dist-release && mkdir -p dist-release
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=""; [ $$os = windows ] && ext=".exe"; \
		name=aiport-$(VERSION)-$$os-$$arch; dir=dist-release/$$name; mkdir -p $$dir; \
		echo "building $$name"; \
		GOOS=$$os GOARCH=$$arch GOPROXY=off go build -trimpath -ldflags "$(LDFLAGS)" -o $$dir/aiport$$ext ./cmd/aiport || exit 1; \
		cp README.md LICENSE NOTICE TRADEMARKS.md THIRD_PARTY.md aiport.example.yaml $$dir/; cp -R licenses $$dir/licenses; \
		if [ $$os = windows ]; then (cd dist-release && zip -qr $$name.zip $$name); \
		else tar -C dist-release -czf dist-release/$$name.tar.gz $$name; fi; \
		rm -rf $$dir; \
	done
	cd dist-release && shasum -a 256 aiport-* > SHA256SUMS && cat SHA256SUMS

# Scan the whole history of every branch for secrets (needs gitleaks: brew install gitleaks).
leaks:
	gitleaks git --config .gitleaks.toml --log-opts="--all" --redact --no-banner .

# Use the repository's git hooks (pre-commit: secrets and your private words).
hooks:
	git config core.hooksPath scripts/hooks
	@echo "hooks installed; add private words (one regex per line) to .git/info/private-patterns"
