@echo off
REM Deploy 5 blockchain nodes for live transaction propagation demo

echo 🌐 BlackHole Blockchain - 5-Node Live Demo
echo ==========================================

cd core\relay-chain\cmd\relay

REM Clean any existing databases
echo [SETUP] Cleaning existing databases...
for /L %%i in (3000,1,3004) do (
    if exist blockchaindb_%%i rmdir /s /q blockchaindb_%%i >nul 2>&1
)

REM Build the blockchain node
echo [BUILD] Building blockchain node...
go build -o blockchain-node.exe .
if errorlevel 1 (
    echo [ERROR] Failed to build blockchain node
    pause
    exit /b 1
)

echo [INFO] ✅ Build successful

REM Start Node 1 (Bootstrap) - Port 3000, API 8080
echo [NODE 1] Starting bootstrap node...
start "Node 1 - Bootstrap" cmd /c "title Node 1 Bootstrap & echo === NODE 1 BOOTSTRAP === & echo P2P: 3000, API: 8080 & echo Starting... & blockchain-node.exe 3000"

REM Wait for bootstrap to start
echo [INFO] Waiting for bootstrap node...
timeout /t 10 /nobreak >nul

REM Start Node 2 - Port 3001, API 8081
echo [NODE 2] Starting node 2...
start "Node 2" cmd /c "title Node 2 & echo === NODE 2 === & echo P2P: 3001, API: 8081 & echo Starting... & blockchain-node.exe 3001"

timeout /t 5 /nobreak >nul

REM Start Node 3 - Port 3002, API 8082
echo [NODE 3] Starting node 3...
start "Node 3" cmd /c "title Node 3 & echo === NODE 3 === & echo P2P: 3002, API: 8082 & echo Starting... & blockchain-node.exe 3002"

timeout /t 5 /nobreak >nul

REM Start Node 4 - Port 3003, API 8083
echo [NODE 4] Starting node 4...
start "Node 4" cmd /c "title Node 4 & echo === NODE 4 === & echo P2P: 3003, API: 8083 & echo Starting... & blockchain-node.exe 3003"

timeout /t 5 /nobreak >nul

REM Start Node 5 - Port 3004, API 8084
echo [NODE 5] Starting node 5...
start "Node 5" cmd /c "title Node 5 & echo === NODE 5 === & echo P2P: 3004, API: 8084 & echo Starting... & blockchain-node.exe 3004"

echo.
echo [SUCCESS] All 5 nodes started!
echo.
echo 🌐 Network Topology:
echo    Node 1 (Bootstrap): P2P 3000, API http://localhost:8080
echo    Node 2:             P2P 3001, API http://localhost:8081
echo    Node 3:             P2P 3002, API http://localhost:8082
echo    Node 4:             P2P 3003, API http://localhost:8083
echo    Node 5:             P2P 3004, API http://localhost:8084
echo.
echo [INFO] Waiting for all nodes to be ready...
timeout /t 15 /nobreak >nul

echo [INFO] Testing node connectivity...
for /L %%i in (8080,1,8084) do (
    echo Testing API on port %%i...
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/health' -Method GET -TimeoutSec 3; Write-Host '✅ Port %%i: ONLINE' } catch { Write-Host '❌ Port %%i: OFFLINE' }"
)

echo.
echo 🎯 Ready for transaction propagation demo!
echo    Use: .\demo-transaction-propagation.bat
echo.
pause
