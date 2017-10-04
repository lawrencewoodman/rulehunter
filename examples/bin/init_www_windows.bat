@echo off
#
#  This command prepares the www directory for the examples
#

set SCRIPT_DIR=%~dp0
set SUPPORT_DIR=%SCRIPT_DIR%..\..\support
set WWW_DIR=%SCRIPT_DIR%..\www

echo Copying www support files to: %WWW_DIR%:
echo.
xcopy %SUPPORT_DIR%\bootstrap\* %WWW_DIR% /e /i /y | find /v "File(s) copied"
xcopy %SUPPORT_DIR%\jquery\* %WWW_DIR% /e /i /y | find /v "File(s) copied"
xcopy %SUPPORT_DIR%\rulehunter\* %WWW_DIR% /e /i /y | find /v "File(s) copied"
xcopy %SUPPORT_DIR%\html5shiv\js\* %WWW_DIR%\js /e /i /y | find /v "File(s) copied"
xcopy %SUPPORT_DIR%\respond\js\* %WWW_DIR%\js /e /i /y | find /v "File(s) copied"
