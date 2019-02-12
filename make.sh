GOOS=linux GOARCH=amd64 go build -o linux/AeroInfoMonitor .
GOOS=windows GOARCH=amd64 go build -o windows/AeroInfoMonitor.exe .
GOOS=darwin GOARCH=amd64 go build -o osx/AeroInfoMonitor .
chmod 755 linux/AeroInfoMonitor
chmod 755 osx/AeroInfoMonitor

