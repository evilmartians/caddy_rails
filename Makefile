.PHONY: build dist test bench clean

COMMIT := $(shell git log --pretty=format:"%h" -n 1)
VERSION := $(shell git tag -l --sort=-version:refname "v*" | head -n1)
LD_FLAGS := "-s -w -X 'github.com/evilmartians/caddy_rails/version.Version=$(VERSION)' -X 'github.com/evilmartians/caddy_rails/version.Commit=$(COMMIT)'"

PLATFORMS = linux darwin freebsd
ARCHITECTURES = amd64 arm64 arm

build:
	go build -ldflags $(LD_FLAGS) -o bin/ ./cmd/...

build-all: clean
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			if [ "$$platform" = "darwin" ] && [ "$$arch" = "arm" ]; then \
				continue; \
			fi; \
			output="dist/caddy-rails-$$platform-$$arch"; \
			echo "Building for $$platform/$$arch..."; \
			env GOOS=$$platform GOARCH=$$arch go build -ldflags $(LD_FLAGS) -o $$output ./cmd/caddy_rails/main.go; \
		done; \
	done

test:
	go test ./...

bench:
	go test -bench=. -benchmem -run=^# ./...

clean:
	rm -rf bin dist
