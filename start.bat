@echo off
setlocal
cd /d "%~dp0"

set "DX_REGION=%~1"
if "%DX_REGION%"=="" set "DX_REGION=cn"
if /I not "%DX_REGION%"=="cn" if /I not "%DX_REGION%"=="global" set "DX_REGION=cn"

echo.
echo Starting DxCloud - region: %DX_REGION%
echo.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\start.ps1" -Region %DX_REGION%

if errorlevel 1 (
  echo.
  echo Startup failed. Check the messages above or run: docker compose logs backend
  pause
  exit /b 1
)

echo.
echo Startup completed. Open http://localhost
pause
endlocal
exit /b 0
