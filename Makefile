.PHONY: node go clean

all: node go

clean:
	rm -fr ui/dist
	rm rice-box.go

node:
	cd ui && npm run build

go:
	rice embed-go
	GOOS=linux go build