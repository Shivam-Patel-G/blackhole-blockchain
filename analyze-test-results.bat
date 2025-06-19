@echo off
REM Multi-Node Test Results Analyzer
REM Analyzes transaction propagation and block creation results

echo 📊 BLACKHOLE BLOCKCHAIN - TEST RESULTS ANALYZER
echo ===============================================

echo [ANALYSIS] Analyzing multi-node test results...
echo.

REM Check if nodes are running
echo [STATUS] Checking node availability...
set online_nodes=0
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/health' -Method GET -TimeoutSec 2; Write-Host '✅ Node %%i: ONLINE'; $env:online_nodes = [int]$env:online_nodes + 1 } catch { Write-Host '❌ Node %%i: OFFLINE' }"
)

echo.
echo [CONSENSUS] Checking network consensus...
echo.

REM Check block height consensus
echo [BLOCKS] Block height verification...
set /a consensus_blocks=0
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/status' -Method GET -TimeoutSec 2; Write-Host 'Node %%i - Block Height:' $response.data.block_height; if ($response.data.block_height -gt 0) { $env:consensus_blocks = [int]$env:consensus_blocks + 1 } } catch { Write-Host 'Node %%i - Block check failed' }"
)

echo.
echo [TRANSACTIONS] Transaction pool verification...
echo.

REM Check transaction pools
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/status' -Method GET -TimeoutSec 2; Write-Host 'Node %%i - Pending Transactions:' $response.data.pending_txs } catch { Write-Host 'Node %%i - Transaction check failed' }"
)

echo.
echo [BALANCES] Balance consistency verification...
echo.

REM Check balance consistency
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/token/balance?address=test-user-1&token=BHX' -Method GET -TimeoutSec 2; Write-Host 'Node %%i - test-user-1 balance:' $response.balance 'BHX' } catch { Write-Host 'Node %%i - Balance check failed' }"
)

echo.
echo [VALIDATORS] Validator network verification...
echo.

REM Check validator network
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/status' -Method GET -TimeoutSec 2; Write-Host 'Node %%i - Validators:' $response.data.validator_count } catch { Write-Host 'Node %%i - Validator check failed' }"
)

echo.
echo [PERFORMANCE] Performance metrics...
echo.

REM Check performance metrics
for /L %%i in (8080,1,8084) do (
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:%%i/api/performance' -Method GET -TimeoutSec 2; Write-Host 'Node %%i - Avg Response:' $response.data.avg_response_time 'ms, Requests:' $response.data.total_requests } catch { Write-Host 'Node %%i - Performance check failed' }"
)

echo.
echo ==========================================
echo TEST RESULTS ANALYSIS COMPLETE
echo ==========================================
echo.

REM Generate summary
echo 📊 SUMMARY REPORT:
echo.

powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:8080/api/status' -Method GET -TimeoutSec 2; Write-Host '🌐 Network Status: ACTIVE'; Write-Host '📦 Total Blocks: ' $response.data.block_height; Write-Host '🔄 Pending Transactions: ' $response.data.pending_txs; Write-Host '👥 Validators: ' $response.data.validator_count } catch { Write-Host '🌐 Network Status: INACTIVE' }"

echo.
echo 🎯 CONSENSUS VERIFICATION:
echo    - All nodes should have the same block height
echo    - All nodes should have consistent transaction pools
echo    - All nodes should show the same balances
echo    - All nodes should have the same validator count
echo.

echo ✅ SUCCESS CRITERIA:
echo    - All 5 nodes online and responding
echo    - Block height synchronized across network
echo    - Transaction propagation working
echo    - Balance consistency maintained
echo    - Validator network stable
echo.

echo 🚀 RECOMMENDATIONS:
echo    - Monitor real-time with real-time-network-monitor.bat
echo    - Submit test transactions with submit-test-transaction.bat
echo    - Run comprehensive test with multi-node-transaction-test.bat
echo.

echo [INFO] Analysis completed successfully!
echo.
pause 