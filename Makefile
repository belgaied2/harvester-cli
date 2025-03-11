build-multiarch:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o harvester-windows-x32.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o harvester-windows-x64.exe .
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o harvester-linux-i386 .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o harvester-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o harvester-linux-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o harvester-darwin-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o harvester-darwin-amd64 .
