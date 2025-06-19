@echo off
REM Real-Time Network Monitor for Multi-Node Blockchain
REM Continuously monitors transaction propagation and block creation

echo 🔍 BLACKHOLE BLOCKCHAIN - REAL-TIME NETWORK MONITOR
echo ===================================================

echo [INFO] Starting real-time network monitoring...
echo [INFO] Monitoring 5 nodes for transaction propagation and block creation
echo [INFO] Press Ctrl+C to stop monitoring
echo.

:monitor_loop
echo.
echo ==========================================
echo TIMESTAMP: %date% %time%
echo ==========================================

REM Check all nodes status
echo [STATUS] Checking all nodes...
for /L %%i in (8080,1,8084) do (
    echo.
    echo --- Node %%i (Port %%i) ---
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/status' -Method GET -TimeoutSec 2; Write-Host '✅ ONLINE - Block:' $response.data.block_height 'Pending:' $response.data.pending_txs 'Validators:' $response.data.validator_count } catch { Write-Host '❌ OFFLINE - Connection failed' }"
)

echo.
echo [TRANSACTIONS] Checking pending transactions across network...
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/status' -Method GET -TimeoutSec 2; if ($response.data.pending_txs -gt 0) { Write-Host 'Node %%i: ' $response.data.pending_txs ' pending transactions' } } catch { }"
)

echo.
echo [BALANCES] Checking test user balances...
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/token/balance?address=test-user-1&token=BHX' -Method GET -TimeoutSec 2; Write-Host 'Node %%i - test-user-1: ' $response.balance ' BHX' } catch { Write-Host 'Node %%i - Balance check failed' }"
)

echo.
echo [HEALTH] Network health check...
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/health' -Method GET -TimeoutSec 2; Write-Host 'Node %%i: ' $response.status ' (Uptime: ' $response.uptime ')' } catch { Write-Host 'Node %%i: HEALTH CHECK FAILED' }"
)

echo.
echo [INFO] Waiting 5 seconds before next check...
echo [INFO] Press Ctrl+C to stop monitoring
timeout /t 5 /nobreak >nul

goto monitor_loop 