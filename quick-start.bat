@echo off
REM Quick Start Script for BlackHole Blockchain
REM This script starts the blockchain without Docker for immediate testing

echo 🚀 BlackHole Blockchain - Quick Start
echo =====================================

echo [INFO] Starting BlackHole Blockchain locally...

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed. Please install Go 1.21+ first.
    echo [INFO] Alternatively, use Docker deployment with: deploy-simple.bat
    pause
    exit /b 1
)

echo [INFO] Go is available

REM Navigate to blockchain directory
cd core\relay-chain\cmd\relay

REM Clean any existing builds
if exist blockchain-node.exe del blockchain-node.exe

REM Build the blockchain node
echo [INFO] Building blockchain node...
go mod tidy
go build -o blockchain-node.exe .

if errorlevel 1 (
    echo [ERROR] Failed to build blockchain node
    echo [INFO] Check for compilation errors above
    pause
    exit /b 1
)

echo [INFO] Build successful

echo [INFO] Starting blockchain node...
echo [INFO] P2P Port: 3000, HTTP API Port: 8080

REM Start the blockchain node with proper port configuration
start "BlackHole Blockchain Node" blockchain-node.exe 3000

REM Wait for startup and check if HTTP server is responding
echo [INFO] Waiting for services to start...
timeout /t 10 /nobreak >nul

REM Check if HTTP server is responding
echo [INFO] Checking HTTP server status...
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8080' -TimeoutSec 10 -UseBasicParsing; Write-Host '✅ HTTP server is responding successfully' } catch { Write-Host '⏳ HTTP server starting up, will be ready shortly' }"

echo.
echo ✅ BlackHole Blockchain is starting!
echo.
echo 🌐 Access Points:
echo    Blockchain Dashboard:   http://localhost:8080
echo    API Endpoint:          http://localhost:8080/api/status
echo    P2P Port:              3000
echo.
echo 🔧 Features Available:
echo    - Enhanced monitoring system
echo    - E2E validation framework  
echo    - Governance simulation
echo    - Load testing capabilities
echo    - Advanced security features
echo.
echo 📊 CLI Commands (in the blockchain window):
echo    status     - Show blockchain status
echo    monitor    - Show monitoring metrics
echo    validate   - Run E2E validation tests
echo    governance - Show governance dashboard
echo    proposal   - Create governance proposal
echo    vote       - Vote on proposals
echo.
echo Press any key to open the blockchain dashboard...
pause >nul

REM Open dashboard
start http://localhost:8080

echo.
echo 🎉 BlackHole Blockchain is now running!
echo    Check the blockchain node window for CLI commands
echo    Dashboard: http://localhost:8080
echo.
pause
