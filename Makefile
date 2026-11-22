TAILWIND ?= ./bin/tailwindcss
TAILWIND_VERSION := v4.3.1
INSTALL_DIR ?= $(HOME)/.local/bin
CSS_IN := internal/site/styles/input.css
CSS_OUT := internal/site/static/styles/app.css

.PHONY: css css-watch generate build dev check-css tailwind

css:
	$(TAILWIND) -i $(CSS_IN) -o $(CSS_OUT) --minify

css-watch:
	$(TAILWIND) -i $(CSS_IN) -o $(CSS_OUT) --watch

generate:
	go tool templ generate

build: css generate
	go build -o bin/hp .
	install -Dm755 bin/hp $(INSTALL_DIR)/hp

# runs air (see .air.toml)
dev:
	go tool air

check-css: css
	git diff --exit-code -- $(CSS_OUT)

tailwind:
	mkdir -p bin
	curl -fsSL -o $(TAILWIND) https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-linux-x64
	chmod +x $(TAILWIND)
