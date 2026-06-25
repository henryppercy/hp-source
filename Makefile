TAILWIND ?= ./bin/tailwindcss
TAILWIND_VERSION := v4.3.1
CSS_IN := internal/site/styles/input.css
CSS_OUT := internal/site/static/styles/app.css

.PHONY: css css-watch build check-css tailwind

css:
	$(TAILWIND) -i $(CSS_IN) -o $(CSS_OUT) --minify

css-watch:
	$(TAILWIND) -i $(CSS_IN) -o $(CSS_OUT) --watch

build: css
	go build -o bin/hp .

check-css: css
	git diff --exit-code -- $(CSS_OUT)

tailwind:
	mkdir -p bin
	curl -fsSL -o $(TAILWIND) https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-linux-x64
	chmod +x $(TAILWIND)
