

set root_dir=%~dp0

if /i "%1"=="help" goto help
if /i "%1"=="--help" goto help
if /i "%1"=="-help" goto help
if /i "%1"=="/help" goto help
if /i "%1"=="?" goto help
if /i "%1"=="-?" goto help
if /i "%1"=="--?" goto help
if /i "%1"=="/?" goto help

@rem Process arguments.
set is_clean=
set is_compile=
set is_test=
set is_install=

:next-arg

if "%1"=="" goto args-done
if /i "%1"=="clean"        set is_clean=1&goto arg-ok
if /i "%1"=="test"         set is_test=1&goto arg-ok
if /i "%1"=="compile"      set is_compile=1&goto arg-ok
if /i "%1"=="install"      set is_compile=1&set is_install=1&goto arg-ok

:arg-ok
shift
goto next-arg
:args-done


echo "GOROOT=%GOROOT%"
echo "GOPATH=%GOPATH%"

@if /i "%ENGINE_PATH%" NEQ "" goto defined_engine_path
@set ENGINE_PATH=%~dp0%..\
:defined_engine_path

@if /i "%PUBLISH_PATH%" NEQ "" goto defined_publish_path
@set PUBLISH_PATH=%~dp0%..\build\
:defined_publish_path


@if not defined test_db_url (
  @set test_db_url=-db.url="host=127.0.0.1 dbname=tpt_models_test user=tpt password=extreme sslmode=disable"
)

@if not defined test_data_db_url (
  @set test_data_db_url=-data_db.url="host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable"
)

@if not defined test_delayed_job_db_url (
  @set test_delayed_job_db_url=-db_url="host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable"
)


@if not defined is_clean goto download_3td_library
for /f "tokens=1 delims=;" %%a in ("%GOPATH%") do (
  del /F /S /Q %%a\src\github.com\
  rmdir /S /Q %%a\src\github.com\
  del /F /S /Q %%a\src\code.google.com\
  rmdir /S /Q %%a\src\code.google.com\
  del /F /S /Q %%a\src\bitbucket.org\
  rmdir /S /Q %%a\src\bitbucket.org\


  del /F /S /Q %%a\pkg\
  rmdir /S /Q %%a\pkg\
)


:download_3td_library
go get github.com/fd/go-shellwords/shellwords
go get github.com/emicklei/go-restful
go get github.com/runner-mei/go-restful
go get github.com/runner-mei/snmpclient
go get github.com/runner-mei/daemontools
go get github.com/runner-mei/delayed_job
go get github.com/garyburd/redigo/redis
go get github.com/grsmv/inflect
go get github.com/ziutek/mymysql
go get github.com/lib/pq
go get github.com/mattn/go-sqlite3
go get code.google.com/p/mahonia
go get code.google.com/p/winsvc
go get bitbucket.org/liamstask/goose

@echo build_install_directory
:build_install_directory
@if not defined is_install goto build_3td_library
@if NOT exist %PUBLISH_PATH% mkdir %PUBLISH_PATH%
@if NOT exist %PUBLISH_PATH%\bin mkdir %PUBLISH_PATH%\bin
@if NOT exist %PUBLISH_PATH%\lib mkdir %PUBLISH_PATH%\lib
@if NOT exist %PUBLISH_PATH%\lib\alerts mkdir %PUBLISH_PATH%\lib\alerts
@if NOT exist %PUBLISH_PATH%\lib\alerts\templates mkdir %PUBLISH_PATH%\lib\alerts\templates
@if NOT exist %PUBLISH_PATH%\conf mkdir %PUBLISH_PATH%\conf

:build_3td_library
@if not defined is_compile goto build_3td_library_ok
for /f "tokens=1 delims=;" %%a in ("%GOPATH%") do (

  cd "%%a\src\github.com\runner-mei\daemontools\daemontools"
  del "*.exe"
  go build
  @if errorlevel 1 goto failed
  @if not defined is_install goto build_delayed_job
  copy daemontools.exe  "%PUBLISH_PATH%\tpt_service_daemon.exe"
  @if errorlevel 1 goto failed

  :build_delayed_job
  cd "..\..\delayed_job\delayed_job"
  del "*.exe"
  go build
  @if errorlevel 1 goto failed
  @if not defined is_install goto build_3td_library_ok
  copy delayed_job.exe  "%PUBLISH_PATH%\bin\tpt_delayed_job.exe"
  @if errorlevel 1 goto failed
)

:build_3td_library_ok
@if not defined is_compile goto install_lua_and_cjson
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

:install_lua_and_cjson
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lib\lua52_amd64.dll" "%ENGINE_PATH%src\lua_binding\"
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lib\cjson_amd64.dll" "%ENGINE_PATH%src\lua_binding\"
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lib\lua52_amd64.dll" "%ENGINE_PATH%src\sampling\"
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\lib\cjson_amd64.dll" "%ENGINE_PATH%src\sampling\"

:test_lua_binding
@if not defined is_test goto build_lua_binding
cd %ENGINE_PATH%src\lua_binding
go test -v
@if errorlevel 1 goto failed
:build_lua_binding

@if not defined is_test goto build_data_store
cd %ENGINE_PATH%src\data_store
go test -v %test_db_url%
@if errorlevel 1 goto failed

cd %ENGINE_PATH%src\data_store\ds
go test -v %test_db_url%
@if errorlevel 1 goto failed

:build_data_store
@if not defined is_compile goto install_data_store
cd %ENGINE_PATH%src\data_store\ds
del "*.exe"
go build
@if errorlevel 1 goto failed

:install_data_store
@if not defined is_install goto test_snmp
copy "ds.exe"  %PUBLISH_PATH%\bin\tpt_ds.exe
@if errorlevel 1 goto failed
xcopy /Y /S /E %ENGINE_PATH%src\data_store\etc\*   %PUBLISH_PATH%\lib\models\
@if errorlevel 1 goto failed

:test_snmp
@if not defined is_test goto build_snmp
cd %ENGINE_PATH%src\snmp
go test -v
@if errorlevel 1 goto failed

:build_snmp

@if not defined is_test goto build_sampling
cd %ENGINE_PATH%src\sampling
go test -v %test_db_url%
@if errorlevel 1 goto failed

:build_sampling
@if not defined is_compile goto install_sampling
cd %ENGINE_PATH%src\sampling\sampling
del "*.exe"
go build
@if errorlevel 1 goto failed

:install_sampling
@if not defined is_install goto test_poller
copy "sampling.exe"  %PUBLISH_PATH%\bin\tpt_sampling.exe
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%src\lua_binding\lib\lua52_amd64.dll" %PUBLISH_PATH%\bin\lua52_amd64.dll
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%src\lua_binding\lib\cjson_amd64.dll" %PUBLISH_PATH%\bin\cjson_amd64.dll
@if errorlevel 1 goto failed
xcopy /Y /S /E "%ENGINE_PATH%src\lua_binding\microlight\*" %PUBLISH_PATH%\lib\microlight\
@if errorlevel 1 goto failed

:test_poller
@if not defined is_test goto build_poller
cd %ENGINE_PATH%src\poller
go test -v %test_db_url% %test_data_db_url% %test_delayed_job_db_url% -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456
@if errorlevel 1 goto failed

:build_poller
@if not defined is_compile goto install_poller
cd %ENGINE_PATH%src\poller\poller
del "*.exe"
go build
@if errorlevel 1 goto failed
:install_poller
@if not defined is_install goto test_carrier
copy "poller.exe" %PUBLISH_PATH%\bin\tpt_poller.exe
@if errorlevel 1 goto failed
xcopy /Y /S /E %ENGINE_PATH%src\poller\templates\*   %PUBLISH_PATH%\lib\alerts\templates\
@if errorlevel 1 goto failed

:test_carrier
@if not defined is_test goto build_carrier
cd %ENGINE_PATH%src\carrier
go test -v  %test_data_db_url% %test_delayed_job_db_url%  -redis=127.0.0.1:9456
@if errorlevel 1 goto failed

:build_carrier
@if not defined is_compile goto install_carrier
cd %ENGINE_PATH%src\carrier\carrier
del "*.exe"
go build
@if errorlevel 1 goto failed

:install_carrier
@if not defined is_install goto build_ok
copy "carrier.exe" %PUBLISH_PATH%\bin\tpt_carrier.exe
@if errorlevel 1 goto failed


REM cd %ENGINE_PATH%src\bridge\discovery_tools
REM go build
REM @if errorlevel 1 goto failed
REM copy "%ENGINE_PATH%src\bridge\discovery_tools\discovery_tools.exe" %~dp0\bin\discovery_tools.exe
REM @if errorlevel 1 goto failed

copy "%ENGINE_PATH%src\autostart_engine.conf" %PUBLISH_PATH%\conf\autostart_engine.conf


:build_ok
@cd %root_dir%
@echo "====================================="
@echo "build success!"
@goto :eof



:help
@echo build.bat clean test compile install
@echo Examples:
@echo   vcbuild.bat compile test     : compile and test
@echo   vcbuild.bat compile          : only compile
@echo   vcbuild.bat test             : only test
@echo   vcbuild.bat install          : compile and install
@echo   vcbuild.bat test install     : test, compile and install
@goto :eof

:failed
@echo "====================================="
@echo "ooooooooo, build failed!"
@cd %root_dir%
@exit /b -1