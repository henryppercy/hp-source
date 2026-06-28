TAILWIND ?= ./bin/tailwindcss
TAILWIND_VERSION := v4.3.1
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

# dev runs air: regenerate templ, rebuild, and restart the server on .go/.templ
# edits. Run `make css-watch` in a separate terminal for CSS (Tailwind's watcher
# must stay in the foreground, so it can't be backgrounded here).
dev:
	go tool air

check-css: css
	git diff --exit-code -- $(CSS_OUT)

tailwind:
	mkdir -p bin
	curl -fsSL -o $(TAILWIND) https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-linux-x64
	chmod +x $(TAILWIND)
