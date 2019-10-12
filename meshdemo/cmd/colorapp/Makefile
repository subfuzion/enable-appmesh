.PHONY: build-dev build-packed clean

build-dev:
	packr2
	go build .

build-packed:
	packr2 clean
	go build .

clean:
	packr2 clean
	rm -f colorapp

