@echo off
setlocal enabledelayedexpansion

cd Sovereign_Map_Federated_Learning

:loop
cls
echo [%date% %time%] - Real-time FL Training Monitor (50 nodes)
echo.
for /f "tokens=*" %%A in ('docker logs sovereign_map_federated_learning-backend-1 --tail 200 2^>^&1 ^| findstr "FL round"') do (
  echo %%A
)
echo.
echo Node Status:
docker ps --filter "label=com.docker.compose.service=node-agent" --format "table {{.Names}}\t{{.Status}}" | find "Up" | find /c "Up"

timeout /t 5 /nobreak
goto loop
