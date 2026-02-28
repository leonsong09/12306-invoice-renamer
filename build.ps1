$ErrorActionPreference = "Stop"

$DIST_DIR = "dist"
$EXE_NAME = "12306-invoice-renamer.exe"

if (-not (Test-Path $DIST_DIR)) {
  New-Item -ItemType Directory -Path $DIST_DIR | Out-Null
}

go build -trimpath -ldflags "-H=windowsgui -s -w" -o (Join-Path $DIST_DIR $EXE_NAME) .\cmd\invoicegui

Write-Host "Built:" (Join-Path $DIST_DIR $EXE_NAME)
