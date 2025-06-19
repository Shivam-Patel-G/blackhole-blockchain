@echo off
REM Quick Test Transaction Submission Script
REM Submit transactions to any node for testing propagation

echo 🚀 BLACKHOLE BLOCKCHAIN - QUICK TRANSACTION TEST
echo ================================================

echo [INFO] Available nodes:
echo    Node 1: http://localhost:8080
echo    Node 2: http://localhost:8081
echo    Node 3: http://localhost:8082
echo    Node 4: http://localhost:8083
echo    Node 5: http://localhost:8084
echo.

set /p node_port="Enter node port (8080-8084): "
set /p amount="Enter amount to transfer: "
set /p recipient="Enter recipient address: "

if "%recipient%"=="" set recipient=test-user-%random%

echo.
echo [SUBMIT] Submitting transaction to Node %node_port%...
echo Transaction: genesis-validator -> %recipient%, Amount: %amount% BHX
echo.

powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%node_port%/api/token/transfer' -Method POST -Body '{\"from\":\"genesis-validator\",\"to\":\"%recipient%\",\"amount\":%amount%,\"token\":\"BHX\"}' -ContentType 'application/json'; Write-Host '✅ SUCCESS: Transaction submitted!'; Write-Host 'TX Hash:' $response.tx_hash; Write-Host 'Status:' $response.status; Write-Host 'Recipient:' $response.to; Write-Host 'Amount:' $response.amount 'BHX' } catch { Write-Host '❌ FAILED: Transaction submission error'; Write-Host 'Status:' $_.Exception.Response.StatusCode; Write-Host 'Details:' $_.Exception.Message }"

echo.
echo [VERIFICATION] Checking transaction propagation...
echo Waiting 5 seconds for propagation...
timeout /t 5 /nobreak >nul

echo.
echo [CHECK] Verifying transaction on all nodes...
for /L %%i in (8080,1,8084) do (
    echo Checking Node %%i...
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/token/balance?address=%recipient%&token=BHX' -Method GET -TimeoutSec 2; Write-Host 'Node %%i - %recipient% balance:' $response.balance 'BHX' } catch { Write-Host 'Node %%i - Balance check failed' }"
)

echo.
echo [INFO] Transaction test completed!
echo [INFO] Use real-time-network-monitor.bat to watch propagation
echo.
pause 