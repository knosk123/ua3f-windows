@echo off
cd /d "%~dp0"
echo ==========================================
echo ua3f-win debug launcher
echo ==========================================
echo.
echo Current directory:
echo %cd%
echo.
echo Launching ua3f-win.exe ...
echo.
ua3f-win.exe
echo.
echo Exit code: %errorlevel%
echo.
pause
