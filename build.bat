go get github.com/runner-mei/go-restful
go get github.com/runner-mei/snmpclient
go get github.com/runner-mei/daemontools
go get github.com/runner-mei/delayed_job
go get github.com/garyburd/redigo/redis
go get github.com/grsmv/inflect
go get github.com/lib/pq
go get github.com/mattn/go-sqlite3
go get code.google.com/p/mahonia

echo "GOROOT=%GOROOT%"
echo "GOPATH=%GOPATH%"

if /i "%ENGINE_PATH%" NEQ "" goto defined_engine_path
set ENGINE_PATH=%~dp0%..\
:defined_engine_path

if /i "%PUBLISH_PATH%" NEQ "" goto defined_publish_path
set PUBLISH_PATH=%~dp0%..\build\
:defined_publish_path


if defined test_db_url (
  echo %test_db_url%
) else (
  set test_db_url=-db.url="host=127.0.0.1 dbname=tpt_models_test user=tpt password=extreme sslmode=disable"
)

if defined test_data_db_url (
  echo %test_data_db_url%
) else (
  set test_data_db_url=-data_db.url="host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable"
)

if defined test_delayed_job_db_url (
  echo %test_delayed_job_db_url%
) else (
  set test_delayed_job_db_url=-db_url="host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable"
)

if NOT exist %PUBLISH_PATH% mkdir %PUBLISH_PATH%
if NOT exist %PUBLISH_PATH%\bin mkdir %PUBLISH_PATH%\bin
if NOT exist %PUBLISH_PATH%\lib mkdir %PUBLISH_PATH%\lib
if NOT exist %PUBLISH_PATH%\lib\alerts mkdir %PUBLISH_PATH%\lib\alerts
if NOT exist %PUBLISH_PATH%\lib\alerts\templates mkdir %PUBLISH_PATH%\lib\alerts\templates
if NOT exist %PUBLISH_PATH%\conf mkdir %PUBLISH_PATH%\conf

for /f "tokens=1 delims=;" %%a in ("%GOPATH%") do (
  cd %%a\src\github.com\runner-mei\daemontools\daemontools
  del "*.exe"
  go build
  @if errorlevel 1 goto failed
  copy daemontools.exe  %PUBLISH_PATH%\tpt_service_daemon.exe


  cd %%a\src\github.com\runner-mei\delayed_job\delayed_job
  del "*.exe"
  go build
  @if errorlevel 1 goto failed
  copy delayed_job.exe  %PUBLISH_PATH%\bin\tpt_delayed_job.exe
)


if NOT exist "%ENGINE_PATH%\src\lua_binding\lib\lua52_amd64.dll" (
  cd %ENGINE_PATH%\src\lua_binding\lua
  mingw32-make mingw_amd64
  @if errorlevel 1 goto failed
  xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lua\src\*.exe" "%ENGINE_PATH%src\lua_binding\lib\"
  xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lua\src\*.dll" "%ENGINE_PATH%src\lua_binding\lib\"
)

if NOT exist "%ENGINE_PATH%\src\lua_binding\lib\cjson_amd64.dll" (
  cd %ENGINE_PATH%\src\lua_binding\lua-cjson
  mingw32-make
  @if errorlevel 1 goto failed
  xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lua-cjson\*.exe" "%ENGINE_PATH%src\lua_binding\lib\"
  xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lua-cjson\*.dll" "%ENGINE_PATH%src\lua_binding\lib\"
)

xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lib\lua52_amd64.dll" "%ENGINE_PATH%src\lua_binding\"
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lib\cjson_amd64.dll" "%ENGINE_PATH%src\lua_binding\"

cd %ENGINE_PATH%src\lua_binding
go test -v
@if errorlevel 1 goto failed

cd %ENGINE_PATH%src\data_store
go test -v %test_db_url%
@if errorlevel 1 goto failed
cd %ENGINE_PATH%src\data_store\ds
go test -v %test_db_url%
@if errorlevel 1 goto failed
del "*.exe"
go build

@if errorlevel 1 goto failed
copy "ds.exe"  %PUBLISH_PATH%\bin\tpt_ds.exe
@if errorlevel 1 goto failed
xcopy /Y /S /E %ENGINE_PATH%src\data_store\etc\*   %PUBLISH_PATH%\lib\models\
@if errorlevel 1 goto failed


cd %ENGINE_PATH%src\snmp
go test -v
@if errorlevel 1 goto failed


cd %ENGINE_PATH%src\sampling
REM go test -v %test_db_url%
REM @if errorlevel 1 goto failed
cd %ENGINE_PATH%src\sampling\sampling
del "*.exe"
go build
@if errorlevel 1 goto failed
copy "sampling.exe"  %PUBLISH_PATH%\bin\tpt_sampling.exe


cd %ENGINE_PATH%src\poller
go test -v %test_db_url% %test_data_db_url% %test_delayed_job_db_url% -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456
@if errorlevel 1 goto failed
cd %ENGINE_PATH%src\poller\poller
del "*.exe"
go build
@if errorlevel 1 goto failed
copy "poller.exe" %PUBLISH_PATH%\bin\tpt_poller.exe
@if errorlevel 1 goto failed


xcopy /Y /S /E %ENGINE_PATH%src\poller\templates\*   %PUBLISH_PATH%\lib\alerts\templates\
@if errorlevel 1 goto failed

cd %ENGINE_PATH%src\carrier
go test -v  %test_data_db_url% %test_delayed_job_db_url%  -redis=127.0.0.1:9456
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
copy "%ENGINE_PATH%src\lua_binding\lib\lua52_amd64.dll" %PUBLISH_PATH%\bin\lua52_amd64.dll
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%src\lua_binding\lib\cjson_amd64.dll" %PUBLISH_PATH%\bin\cjson_amd64.dll
@if errorlevel 1 goto failed
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\microlight\*" %PUBLISH_PATH%\lib\microlight\
@if errorlevel 1 goto failed

copy "%ENGINE_PATH%src\autostart_engine.conf" %PUBLISH_PATH%\conf\autostart_engine.conf


@echo "====================================="
@echo "build success!"
@goto :eof

:failed
@echo "====================================="
@echo "ooooooooo, build failed!"
cd %~dp0
exit /b -1