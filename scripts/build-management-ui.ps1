$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$uiDir = Join-Path $root 'frontend\management-center'
$output = Join-Path $root 'internal\managementasset\management.html'

if (-not (Test-Path (Join-Path $uiDir 'node_modules'))) {
  Push-Location $uiDir
  try {
    npm ci
  } finally {
    Pop-Location
  }
}

Push-Location $uiDir
try {
  npm run build
} finally {
  Pop-Location
}

Copy-Item -Path (Join-Path $uiDir 'dist\index.html') -Destination $output -Force
Write-Host "management UI synced to $output"
