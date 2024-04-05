.PHONY: install
install:
	go build
	cp -f findgrep ~/.local/bin
