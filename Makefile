VERSION := $$(cat VERSION)
all: build strip
build: impromon_linux_amd64 impromon_linux_arm64 impromon_linux_arm impromon_linux_arm64 impromon_mac_amd64 impromon_mac_arm64 impromon_mac_amd64 impromon_windows_amd64.exe impromon_windows_arm64.exe
impromon_linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/impromon_linux_amd64_$(VERSION)
impromon_windows_amd64.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/impromon_win_amd64_$(VERSION).exe
impromon_windows_arm64.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o build/impromon_win_arm64_$(VERSION).exe
impromon_mac_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/impromon_mac_amd64_$(VERSION)
impromon_mac_arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o build/impromon_mac_arm64_$(VERSION)
impromon_linux_arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o build/impromon_linux_arm_$(VERSION)
impromon_linux_arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/impromon_linux_arm64_$(VERSION)

strip:
	find build -name "build/*" -exec strip {} \;

zip:
	cd build
	zip -r build/impromon.zip build/*

prod: build strip zip

clean:
	rm build/*