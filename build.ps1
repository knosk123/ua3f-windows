param()

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $projectRoot

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " ua3f-win build script" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

$goVersion = & go version 2>$null
if (-not $goVersion) {
    Write-Host "[X] Go was not found. Install Go 1.25+ first." -ForegroundColor Red
    exit 1
}
Write-Host "[+] $goVersion"

$dllPath = Join-Path $projectRoot "windivert_amd64.dll"
$sysPath = Join-Path $projectRoot "windivert_amd64.sys"
if (-not (Test-Path $dllPath) -or -not (Test-Path $sysPath)) {
    Write-Host "[X] Missing WinDivert files in project root." -ForegroundColor Red
    Write-Host "    Required: windivert_amd64.dll and windivert_amd64.sys" -ForegroundColor Yellow
    exit 1
}
Write-Host "[+] WinDivert files found"

if (-not $env:GOPROXY) {
    $env:GOPROXY = "https://goproxy.cn,direct"
}
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "[*] Running go mod tidy ..."
& go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "[X] go mod tidy failed" -ForegroundColor Red
    exit 1
}

Write-Host "[*] Building ua3f-win.exe ..."
& go build -ldflags "-s -w" -o "$projectRoot\ua3f-win.exe"
if ($LASTEXITCODE -ne 0) {
    Write-Host "[X] go build failed" -ForegroundColor Red
    exit 1
}

$fileInfo = Get-Item "$projectRoot\ua3f-win.exe"
$sizeKB = [math]::Round($fileInfo.Length / 1KB, 1)

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host " build finished" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host " output : ua3f-win.exe ($sizeKB KB)" -ForegroundColor White
Write-Host " default: .\ua3f-win.exe" -ForegroundColor White
Write-Host " preset : .\ua3f-win.exe -ua pc" -ForegroundColor White
Write-Host " debug  : .\ua3f-win.exe -ua wechat -log debug" -ForegroundColor White
Write-Host ""
