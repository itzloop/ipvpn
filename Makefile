.PHONY: build clean

deploy_pi: build_linux_arm64
	scp -i ~/.ssh/temparvan ./build/ipvpn-linux_arm64 pi:

build: build_linux build_linux_arm64 build_win build_darwin build_darwin_arm64

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/ipvpn-linux_amd64 main.go

build_darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./build/ipvpn-darwin_amd6 main.go

build_win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/ipvpn-win_amd64.exe main.go

build_linux_arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./build/ipvpn-linux_arm64 main.go

build_darwin_arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./build/ipvpn-darwin_arm64 main.go

clean:
	rm -r ./build
