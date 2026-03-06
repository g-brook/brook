@echo off
REM -------------------------------
REM Start the brook service
REM -------------------------------

echo Starting brook service...
REM Change the path below to your actual brook.exe location if needed
cd /d "%~dp0"
.\brook-cli.exe

REM Keep the window open to view logs
pause