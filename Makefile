.PHONY: vendor

help:
	# "make [vendor|front|back|bundle|dev]"

bundle: clean vendor test front back

clean:
	@rm -rf ~/bin/gps-frontend ~/bin/gps-gatewayd cmd/frontend/pkged.go

vendor:
	@go mod tidy
	@go mod vendor
	@go mod download

test:
	@go fmt ./...
	@go test -v pkg/gpshome/*

front:
	@go generate cmd/frontend/*.go
	@go build -o ~/bin/gps-frontend cmd/frontend/*go

back:
	@go build -o ~/bin/gps-gatewayd cmd/gatewayd/*go

push:
	@scp ~/bin/gps-frontend gps.host:
	@scp ~/bin/gps-gatewayd gps.host:

dev-front:
	@go build -o ~/bin/gps-frontend cmd/frontend/*go
	@gps-frontend

