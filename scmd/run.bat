@echo off
REM -------------------------------
REM Start the brook service and check if it started
REM -------------------------------

echo Starting brook service...
cd /d "%~dp0"

REM Start brook in a separate window so batch can continue
start "" .\brook-sev.exe

REM Wait a few seconds for the service to initialize
timeout /t 3 > nul

REM Check if brook-sev.exe is running
tasklist /FI "IMAGENAME eq brook-sev.exe" | findstr /I "brook-sev.exe" > nul
if %ERRORLEVEL%==0 (
    echo Brook service started successfully.
) else (
    echo Failed to start Brook service.
)

REM Keep the window open to view logs
pause