go get github.com/runner-mei/go-restful
go get github.com/runner-mei/snmpclient
go get github.com/runner-mei/daemontools
go get github.com/runner-mei/delayed_job
go get github.com/garyburd/redigo/redis
go get github.com/grsmv/inflect
go get github.com/lib/pq
go get github.com/mattn/go-sqlite3
go get code.google.com/p/mahonia

if /i "%ENGINE_PATH%" NEQ "" goto defined_engine_path
set ENGINE_PATH=%~dp0%..\
:defined_engine_path

if /i "%PUBLISH_PATH%" NEQ "" goto defined_publish_path
set PUBLISH_PATH=%~dp0%..\build\
:defined_publish_path


if NOT exist %PUBLISH_PATH% mkdir %PUBLISH_PATH%
if NOT exist %PUBLISH_PATH%\bin mkdir %PUBLISH_PATH%\bin
if NOT exist %PUBLISH_PATH%\lib mkdir %PUBLISH_PATH%\lib


for /f "tokens=1 delims=;" %%a in ("%GOPATH%") do (
  cd %%a\src\github.com\runner-mei\daemontools\daemontools
  del "*.exe"
  go build
  @if errorlevel 1 goto failed
  copy daemontools.exe  %PUBLISH_PATH%\bin\tpt_daemontools.exe


  cd %%a\src\github.com\runner-mei\delayed_job\delayed_job
  del "*.exe"
  go build
  @if errorlevel 1 goto failed
  copy delayed_job.exe  %PUBLISH_PATH%\bin\tpt_delayed_job.exe
)

cd %ENGINE_PATH%src\data_store\ds
del "*.exe"
go build

@if errorlevel 1 goto failed
copy "ds.exe"  %PUBLISH_PATH%\bin\tpt_ds.exe
@if errorlevel 1 goto failed
xcopy /Y /S /E %ENGINE_PATH%src\data_store\etc\*   %PUBLISH_PATH%\lib\models\


cd %ENGINE_PATH%src\sampling\sampling
del "*.exe"
go build
@if errorlevel 1 goto failed
copy "sampling.exe"  %PUBLISH_PATH%\bin\tpt_sampling.exe


cd %ENGINE_PATH%src\poller\poller
del "*.exe"
go build
@if errorlevel 1 goto failed
copy "poller.exe" %PUBLISH_PATH%\bin\tpt_poller.exe
@if errorlevel 1 goto failed


cd %ENGINE_PATH%src\carrier\carrier
del "*.exe"
go build
@if errorlevel 1 goto failed
copy "carrier.exe" %PUBLISH_PATH%\bin\tpt_carrier.exe
@if errorlevel 1 goto failed


REM cd %ENGINE_PATH%src\bridge\discovery_tools
REM go build
REM @if errorlevel 1 goto failed
REM copy "%ENGINE_PATH%src\bridge\discovery_tools\discovery_tools.exe" %~dp0\bin\discovery_tools.exe
REM @if errorlevel 1 goto failed

cd %~dp0
copy "%ENGINE_PATH%src\lua_binding\lib\lua52.dll" %PUBLISH_PATH%\bin\lua52.dll
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%src\lua_binding\lib\cjson.dll" %PUBLISH_PATH%\bin\cjson.dll
@if errorlevel 1 goto failed
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\microlight\*" %PUBLISH_PATH%\lib\microlight\
@if errorlevel 1 goto failed

@goto :eof

:failed
@echo "±‡“Î ß∞‹"
cd %~dp0