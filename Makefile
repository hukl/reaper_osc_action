.DEFAULT_GOAL := build

.PHONY:fmt vet build test plugin
fmt:
	go fmt

vet: fmt
	go vet

build: vet
	GOOS=darwin GOARCH=amd64 go build -o org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_macos-amd64
	GOOS=darwin GOARCH=arm64 go build -o org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_macos-arm64
	lipo -create -output org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_macos-universal org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_macos-amd64 org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_macos-arm64
	rm org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_macos-amd64
	rm org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_macos-arm64
	GOOS=windows GOARCH=amd64 go build -o org.smyck.reaper-osc-action.sdPlugin/reaper_osc_action_win-amd64

test:
	go test ./...

plugin:
	# Requires fd and sd executables
	# https://github.com/sharkdp/fd
	# https://github.com/elgatosf/cli
	fd -H -I '.DS_Store' -x rm -f
	sd pack -f org.smyck.reaper-osc-action.sdPlugin
