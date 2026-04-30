param(
    [switch]$Build,
    [switch]$WithRedis
)

$ErrorActionPreference = "Stop"

$composeArgs = @()
if ($WithRedis) {
    $composeArgs += @("--profile", "optional")
}

$upArgs = @("up", "-d")
if ($Build) {
    $upArgs += "--build"
}

$services = @("postgres", "api", "web")
if ($WithRedis) {
    $services += "redis"
}

docker compose @composeArgs @upArgs @services

Write-Host ""
Write-Host "DevsVault esta levantado:"
Write-Host "- Web:    http://localhost:3000"
Write-Host "- API:    http://localhost:8080"
Write-Host "- Health: http://localhost:8080/healthz"
Write-Host ""
Write-Host "Login de desarrollo sugerido:"
Write-Host "- API URL: http://localhost:8080"
Write-Host "- Subject: admin@example.local"
Write-Host "- Actor:   user"
Write-Host ""
Write-Host "CLI local:"
Write-Host "go run ./apps/cli/cmd/devsvault login --url http://localhost:8080 --subject admin@example.local --type user"
