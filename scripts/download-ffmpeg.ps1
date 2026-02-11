# download-ffmpeg.ps1
# Downloads FFmpeg essentials build and extracts ffmpeg.exe
# to build/windows/installer/resources/ for NSIS bundling.

param(
    [string]$OutputDir = "$PSScriptRoot\..\build\windows\installer\resources"
)

$ErrorActionPreference = "Stop"

# FFmpeg release URL (gyan.dev essentials build - widely used, stable)
$FfmpegUrl = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
$TempZip = Join-Path $env:TEMP "ffmpeg-essentials.zip"
$TempExtract = Join-Path $env:TEMP "ffmpeg-extract"

# Ensure output directory exists
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

$OutputDir = Resolve-Path $OutputDir

Write-Host "Downloading FFmpeg essentials build..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $FfmpegUrl -OutFile $TempZip -UseBasicParsing

Write-Host "Extracting..." -ForegroundColor Cyan
if (Test-Path $TempExtract) {
    Remove-Item -Recurse -Force $TempExtract
}
Expand-Archive -Path $TempZip -DestinationPath $TempExtract -Force

# Find ffmpeg.exe inside the extracted directory
$FfmpegExe = Get-ChildItem -Path $TempExtract -Recurse -Filter "ffmpeg.exe" |
    Where-Object { $_.DirectoryName -like "*\bin" } |
    Select-Object -First 1

if (-not $FfmpegExe) {
    Write-Error "ffmpeg.exe not found in the downloaded archive"
    exit 1
}

$Destination = Join-Path $OutputDir "ffmpeg.exe"
Copy-Item -Path $FfmpegExe.FullName -Destination $Destination -Force

Write-Host "FFmpeg copied to: $Destination" -ForegroundColor Green
Write-Host "Size: $([math]::Round($FfmpegExe.Length / 1MB, 2)) MB" -ForegroundColor Green

# Cleanup
Remove-Item -Force $TempZip -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force $TempExtract -ErrorAction SilentlyContinue

Write-Host "Done! FFmpeg is ready for NSIS bundling." -ForegroundColor Green
