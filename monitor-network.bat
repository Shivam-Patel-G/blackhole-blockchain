@echo off
REM Real-time Multi-Node Network Monitoring Tool
REM Provides continuous monitoring of all blockchain nodes

echo 🔍 BlackHole Blockchain - Network Monitor
echo ==========================================

:MONITOR_LOOP

cls
echo 🔍 BlackHole Blockchain - Network Monitor
echo ==========================================
echo Timestamp: %date% %time%
echo.

REM Check if nodes are running
set node_count=0
for /f %%i in ('tasklist /fi "imagename eq blockchain-node.exe" /fo csv ^| find /c /v ""') do set /a node_count=%%i-1

if %node_count% LEQ 0 (
    echo ❌ No blockchain nodes detected
    echo Please run deploy-multi-node.bat first
    echo.
    goto MONITOR_END
)

echo 📊 Active Nodes: %node_count%
echo.

echo ==========================================
echo NODE STATUS OVERVIEW
echo ==========================================

REM Check each node's status
for /L %%i in (8080,1,8089) do (
    set /a node_id=%%i-8080+1
    powershell -Command "$ErrorActionPreference='SilentlyContinue'; try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/status' -Method GET -TimeoutSec 2; $status = if ($response.success) { 'ONLINE' } else { 'ERROR' }; Write-Host ('Node {0:D2}: {1} | Height: {2} | Pending: {3} | Validators: {4}' -f !node_id!, $status, $response.data.block_height, $response.data.pending_txs, $response.data.validator_count) } catch { Write-Host ('Node {0:D2}: OFFLINE' -f !node_id!) }"
)

echo.
echo ==========================================
echo NETWORK CONSENSUS STATUS
echo ==========================================

REM Get block heights from all nodes
echo Block Height Distribution:
powershell -Command "$heights = @{}; for ($i=8080; $i -le 8089; $i++) { try { $response = Invoke-RestMethod -Uri \"http://localhost:$i/api/status\" -Method GET -TimeoutSec 2; $height = $response.data.block_height; if ($heights.ContainsKey($height)) { $heights[$height]++ } else { $heights[$height] = 1 } } catch { } }; $heights.GetEnumerator() | Sort-Object Key | ForEach-Object { Write-Host \"  Height $($_.Key): $($_.Value) nodes\" }"

echo.
echo Validator Count Distribution:
powershell -Command "$validators = @{}; for ($i=8080; $i -le 8089; $i++) { try { $response = Invoke-RestMethod -Uri \"http://localhost:$i/api/status\" -Method GET -TimeoutSec 2; $count = $response.data.validator_count; if ($validators.ContainsKey($count)) { $validators[$count]++ } else { $validators[$count] = 1 } } catch { } }; $validators.GetEnumerator() | Sort-Object Key | ForEach-Object { Write-Host \"  $($_.Key) validators: $($_.Value) nodes\" }"

echo.
echo ==========================================
echo TRANSACTION POOL STATUS
echo ==========================================

REM Check pending transactions across nodes
powershell -Command "$totalPending = 0; $nodesPending = 0; for ($i=8080; $i -le 8089; $i++) { try { $response = Invoke-RestMethod -Uri \"http://localhost:$i/api/status\" -Method GET -TimeoutSec 2; $pending = $response.data.pending_txs; $totalPending += $pending; if ($pending -gt 0) { $nodesPending++ } } catch { } }; Write-Host \"Total Pending Transactions: $totalPending\"; Write-Host \"Nodes with Pending Transactions: $nodesPending\""

echo.
echo ==========================================
echo NETWORK HEALTH INDICATORS
echo ==========================================

REM Calculate network health metrics
powershell -Command "$online = 0; $consensus = $true; $lastHeight = -1; for ($i=8080; $i -le 8089; $i++) { try { $response = Invoke-RestMethod -Uri \"http://localhost:$i/api/status\" -Method GET -TimeoutSec 2; $online++; $height = $response.data.block_height; if ($lastHeight -eq -1) { $lastHeight = $height } elseif ([Math]::Abs($height - $lastHeight) -gt 1) { $consensus = $false } } catch { } }; $healthScore = [Math]::Round(($online / 10) * 100); Write-Host \"Network Uptime: $online/10 nodes ($healthScore%)\"; $consensusStatus = if ($consensus) { 'GOOD' } else { 'DIVERGED' }; Write-Host \"Consensus Status: $consensusStatus\""

echo.
echo ==========================================
echo REAL-TIME COMMANDS
echo ==========================================
echo Press 'Q' to quit, 'T' to run tests, 'R' to refresh now
echo Auto-refresh in 10 seconds...

REM Wait for user input or timeout
choice /c QTR /t 10 /d R /n >nul

if errorlevel 3 goto MONITOR_LOOP
if errorlevel 2 goto RUN_TESTS  
if errorlevel 1 goto MONITOR_END

goto MONITOR_LOOP

:RUN_TESTS
echo.
echo 🧪 Running quick network tests...
call test-multi-node.bat
pause
goto MONITOR_LOOP

:MONITOR_END
echo.
echo Network monitoring stopped.
pause
