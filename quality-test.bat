@echo off
setlocal
cd /d "%~dp0"
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0tools\quality-acceptance.ps1"
if errorlevel 1 (
  echo.
  echo Quality test failed.
  pause
  exit /b 1
)
pause
endlocal
exit /b 0
