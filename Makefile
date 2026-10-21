.PHONY: build run gallery gallery-build test race vet smoke smoke-compare package qa-setup qa-start qa-restart qa-watch qa-stop qa-status qa-verify

build:
	go build -trimpath -o dist/native/rec-tablet ./cmd/rec-tablet

run: build
	./dist/native/rec-tablet $(ARGS)

gallery-build:
	go build -trimpath -o dist/native/ui-gallery ./cmd/ui-gallery

gallery: gallery-build
	./dist/native/ui-gallery $(ARGS)

test:
	go test -tags ci ./cmd/... ./src/... ./next/...

race:
	go test -race -tags ci ./cmd/... ./src/... ./next/...

vet:
	go vet -tags ci ./cmd/... ./src/... ./next/...

smoke:
	go build -tags ci -o dist/native/rec-tablet-headless ./cmd/rec-tablet
	mkdir -p dist/native/smoke-cache
	XDG_CACHE_HOME="$(CURDIR)/dist/native/smoke-cache" dist/native/rec-tablet-headless --database "$(CURDIR)/dist/native/smoke.sqlite3" --smoke-test --screenshot "$(CURDIR)/dist/native/smoke.png"

smoke-compare:
	go build -tags ci -o dist/native/rec-tablet-headless ./cmd/rec-tablet
	mkdir -p dist/native/smoke-cache
	XDG_CACHE_HOME="$(CURDIR)/dist/native/smoke-cache" dist/native/rec-tablet-headless --ui=next --database "$(CURDIR)/dist/native/smoke-next.sqlite3" --smoke-test --screenshot "$(CURDIR)/dist/native/smoke-next.png"
	XDG_CACHE_HOME="$(CURDIR)/dist/native/smoke-cache" dist/native/rec-tablet-headless --ui=legacy --database "$(CURDIR)/dist/native/smoke-legacy.sqlite3" --smoke-test --screenshot "$(CURDIR)/dist/native/smoke-legacy.png"

package: build
	mkdir -p dist/native/package/usr/bin dist/native/package/usr/share/applications dist/native/package/usr/share/icons/hicolor/scalable/apps dist/native/package/usr/share/doc/rec-tablet
	cp dist/native/rec-tablet dist/native/package/usr/bin/
	cp packaging/rec-tablet.desktop dist/native/package/usr/share/applications/
	cp packaging/rec-tablet.svg dist/native/package/usr/share/icons/hicolor/scalable/apps/
	cp README.md docs/native-poc.md docs/device-trial.md docs/iphone-qa.md docs/code-guide.html dist/native/package/usr/share/doc/rec-tablet/
	tar -C dist/native/package -czf dist/native/rec-tablet-linux-$$(go env GOARCH).tar.gz usr

qa-setup:
	./scripts/iphone-qa setup

qa-start:
	./scripts/iphone-qa start

qa-restart:
	./scripts/iphone-qa restart

qa-watch:
	./scripts/iphone-qa watch

qa-stop:
	./scripts/iphone-qa stop

qa-status:
	./scripts/iphone-qa status

qa-verify:
	./scripts/iphone-qa verify
