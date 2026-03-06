#!/usr/bin/env pwsh
# PowerShell скрипт для создания папок

$folders = @(
    "01-atomic-inventory",
    "02-anti-bruteforce",
    "03-worker-pool",
    "04-idempotency",
    "05-rate-limiter",
    "06-multilayer-cache",
    "07-bff",
    "08-gateway",
    "09-data-mocker",
    "10-readwrite-splitter",
    "11-sharder",
    "12-multitenancy",
    "13-chat-engine",
    "14-leaderboard",
    "15-saga",
    "16-circuit-breaker",
    "17-feature-toggle",
    "18-log-aggregator",
    "19-metrics-exporter",
    "20-dashboard",
    "21-kafka-exactly-once",
    "22-event-sourcing",
    "23-job-scheduler",
    "24-cdc",
    "25-tcp-proxy",
    "26-zero-copy",
    "27-binary-protocol",
    "28-distributed-lock",
    "29-merkle-tree",
    "30-wallet"
)

Write-Host "📁 Creating project folders..." -ForegroundColor Cyan

foreach ($folder in $folders) {
    New-Item -ItemType Directory -Path $folder -Force | Out-Null
    Write-Host "  ✅ Created: $folder" -ForegroundColor Green
}

Write-Host "🎉 All 30 folders created successfully!" -ForegroundColor Yellow
