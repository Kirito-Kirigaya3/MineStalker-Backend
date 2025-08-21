@echo off

echo Building the Go project...

REM Ensure Go modules are tidy
go mod tidy

REM Build the executable
go build -o minestalker.exe

REM Check if the build succeeded
IF %ERRORLEVEL% EQU 0 (
    echo Build succeeded. Running the executable...
    minestalker.exe
) ELSE (
    echo Build failed.
    exit /b 1
)
