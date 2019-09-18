@ECHO OFF
SETLOCAL EnableDelayedExpansion

set ROOT=sysl2\sysl\tests
for /R %ROOT% %%f in (*.sysl) do (
	%GOPATH%\bin\sysl.exe --log debug pb --mode textpb --root %ROOT% -o %ROOT%\%%~nf.win.txt  /%%~nf.sysl || exit /b !errorlevel!
)

del sysl2\sysl\tests\*.win.txt

%GOPATH%\bin\sysl.exe --version
%GOPATH%\bin\sysl.exe --info
