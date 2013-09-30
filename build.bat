

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
@set ENGINE_PATH=%~dp0%..
:defined_engine_path

@if /i "%PUBLISH_PATH%" NEQ "" goto defined_publish_path
@set PUBLISH_PATH=%~dp0%..\build
:defined_publish_path

echo "ENGINE_PATH=%ENGINE_PATH%"
echo "PUBLISH_PATH=%PUBLISH_PATH%"

@if not defined test_db_url (
  @set test_db_url=-db.url="host=127.0.0.1 dbname=tpt_models_test user=tpt password=extreme sslmode=disable"
)

@if not defined test_data_db_url (
  @set test_data_db_url=-data_db.url="host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable"
)

@if not defined test_data_db_url_mysql (
  @set test_data_db_url_mysql=-data_db.driver=mysql -data_db.url="tpt:extreme@tcp(localhost:3306)/tpt_data?autocommit=true&parseTime=true"
)

@if not defined test_data_db_url_mssql (
  @set test_delayed_job_db_url_mssql=-data_db.driver=odbc_with_mssql -data_db.url="dsn=tpt;uid=tpt;pwd=extreme"
)

@if not defined test_delayed_job_db_url (
  @set test_delayed_job_db_url=-db_url="host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable"
)

@if not defined test_delayed_job_db_url_mysql (
  @set test_delayed_job_db_url_mysql=-db_drv=mysql -db_url="tpt:extreme@tcp(localhost:3306)/tpt_data?autocommit=true&parseTime=true"
)

@if not defined test_delayed_job_db_url_mssql (
  @set test_delayed_job_db_url_mssql=-db_drv=odbc_with_mssql -db_url="dsn=tpt;uid=tpt;pwd=extreme"
)

@if not defined test_delayed_job_db_url_for_test (
  @set test_delayed_job_db_url_for_test=-notification_db_drv=postgres -notification_db_url="host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable"
)

@if not defined test_delayed_job_db_url_mysql_for_test (
  @set test_delayed_job_db_url_mysql_for_test=-notification_db_drv=mysql -notification_db_url="tpt:extreme@tcp(localhost:3306)/tpt_data?autocommit=true&parseTime=true"
)

@if not defined test_delayed_job_db_url_mssql_for_test (
  @set test_delayed_job_db_url_mssql_for_test=-notification_db_drv=odbc_with_mssql -notification_db_url="dsn=tpt;uid=tpt;pwd=extreme"
)

@if not defined test_delayed_job_db_url_oracle_for_test (
  @set test_delayed_job_db_url_oracle_for_test=-notification_db_drv=odbc_with_oracle -notification_db_url="DSN=tpt_oracle;UID=system;PWD=123456"
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
go get code.google.com/p/mahonia
go get code.google.com/p/winsvc

go get bitbucket.org/runner_mei/goose
go get bitbucket.org/runner_mei/goose/goose
go get bitbucket.org/phiggins/go-db2-cli
go get bitbucket.org/rj/odbc3-go
go get bitbucket.org/miquella/mgodbc
go get code.google.com/p/odbc
go get github.com/LukeMauldin/lodbc
go get github.com/weigj/go-odbc
go get github.com/ziutek/mymysql/mysql
go get github.com/ziutek/mymysql/native
go get github.com/ziutek/mymysql/godrv
go get github.com/go-sql-driver/mysql
go get github.com/lib/pq
go get github.com/mattn/go-sqlite3
go get github.com/mattn/go-oci8
go get github.com/mattn/go-adodb
go get github.com/wendal/go-oci8
go get github.com/tgulacsi/goracle/oracle
go get github.com/tgulacsi/goracle/godrv



:build_install_directory
@if not defined is_install goto build_3td_library
@if NOT exist %PUBLISH_PATH% mkdir %PUBLISH_PATH%
@if NOT exist %PUBLISH_PATH%\bin mkdir %PUBLISH_PATH%\bin
@if NOT exist %PUBLISH_PATH%\lib mkdir %PUBLISH_PATH%\lib
@if NOT exist %PUBLISH_PATH%\lib\alerts mkdir %PUBLISH_PATH%\lib\alerts
@if NOT exist %PUBLISH_PATH%\lib\alerts\templates mkdir %PUBLISH_PATH%\lib\alerts\templates
@if NOT exist %PUBLISH_PATH%\conf mkdir %PUBLISH_PATH%\conf
@if NOT exist %PUBLISH_PATH%\tools mkdir %PUBLISH_PATH%\tools
@if NOT exist %PUBLISH_PATH%\lib\data-migrations mkdir %PUBLISH_PATH%\lib\data-migrations

:build_3td_library
@if not defined is_compile goto build_3td_library_ok
for /f "tokens=1 delims=;" %%a in ("%GOPATH%") do (
  cd "%%a\src\github.com\runner-mei\daemontools\daemontools"
  del "*.exe"
  go build
  @if errorlevel 1 goto failed
  @if not defined is_install goto build_3td_library_ok
  copy daemontools.exe  "%PUBLISH_PATH%\tpt_service_daemon.exe"
  @if errorlevel 1 goto failed
)

:build_3td_library_ok
cd %ENGINE_PATH%\src\goose
del "*.exe"
go build
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%\src\goose\goose.exe"  "%PUBLISH_PATH%\tools\goose.exe"

cd %ENGINE_PATH%\src\license\sc
del "*.exe"
go build
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%\src\license\sc\sc.exe"  "%PUBLISH_PATH%\tools\tpt_sc.exe"

cd %ENGINE_PATH%\src\delayed_job
del "*.exe"
go build
@if errorlevel 1 goto failed
copy delayed_job.exe  "%PUBLISH_PATH%\bin\tpt_delayed_job.exe"
@if errorlevel 1 goto failed

:build_lua
@if not defined is_clean goto build_lua_and_cjson
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lib\*.dll"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lib\*.exe"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lib\*.so"

del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua\src\*.a"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua\src\*.o"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua\src\*.obj"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua\src\*.lib"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua\src\*.so"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua\src\*.dll"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua\src\*.exe"

del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.a"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.o"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.obj"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.lib"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.so"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.dll"
del /S /Q /F "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.exe"

:build_lua_and_cjson
@if not defined is_compile goto install_lua_and_cjson
if NOT exist "%ENGINE_PATH%\src\lua_binding\lib\lua52_amd64.dll" (
  cd %ENGINE_PATH%\src\lua_binding\lua
  mingw32-make mingw_amd64
  @if errorlevel 1 goto failed
  xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lua\src\*.exe" "%ENGINE_PATH%\src\lua_binding\lib\"
  xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lua\src\*.dll" "%ENGINE_PATH%\src\lua_binding\lib\"
)

if NOT exist "%ENGINE_PATH%\src\lua_binding\lib\cjson_amd64.dll" (
  cd %ENGINE_PATH%\src\lua_binding\lua-cjson
  mingw32-make
  @if errorlevel 1 goto failed
  xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.exe" "%ENGINE_PATH%\src\lua_binding\lib\"
  xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lua-cjson\*.dll" "%ENGINE_PATH%\src\lua_binding\lib\"
)

:install_lua_and_cjson
xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lib\lua52_amd64.dll" "%ENGINE_PATH%\src\lua_binding\"
xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lib\cjson_amd64.dll" "%ENGINE_PATH%\src\lua_binding\"

:test_lua_binding
@if not defined is_test goto build_lua_binding
cd %ENGINE_PATH%\src\lua_binding
go test -v
@if errorlevel 1 goto failed
:build_lua_binding

@if not defined is_test goto build_data_store
cd %ENGINE_PATH%\src\data_store
go test -v %test_db_url%
@if errorlevel 1 goto failed

cd %ENGINE_PATH%\src\data_store\ds
go test -v %test_db_url%
@if errorlevel 1 goto failed

:build_data_store
@if not defined is_compile goto install_data_store
cd %ENGINE_PATH%\src\data_store\ds
del "*.exe"
go build
@if errorlevel 1 goto failed

:install_data_store
@if not defined is_install goto test_snmp
copy "%ENGINE_PATH%\src\data_store\ds\ds.exe"  "%PUBLISH_PATH%\bin\tpt_ds.exe"
@if errorlevel 1 goto failed
xcopy /Y /S /E "%ENGINE_PATH%\src\data_store\etc\*"   %PUBLISH_PATH%\lib\models\
@if errorlevel 1 goto failed

:test_snmp
@if not defined is_test goto build_snmp
cd %ENGINE_PATH%\src\snmp
go test -v
@if errorlevel 1 goto failed

:build_snmp

:test_sampling
xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lib\lua52_amd64.dll" "%ENGINE_PATH%\src\sampling\"
xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\lib\cjson_amd64.dll" "%ENGINE_PATH%\src\sampling\"
@if not defined is_test goto build_sampling
cd %ENGINE_PATH%\src\sampling
go test -v %test_db_url%
@if errorlevel 1 goto failed

:build_sampling
@if not defined is_compile goto install_sampling
cd %ENGINE_PATH%\src\sampling\sampling
del "*.exe"
go build
@if errorlevel 1 goto failed
cd %ENGINE_PATH%\src\sampling\snmptools
del "*.exe"
go build
@if errorlevel 1 goto failed

:install_sampling
@if not defined is_install goto test_carrier
copy "%ENGINE_PATH%\src\sampling\sampling\sampling.exe"  "%PUBLISH_PATH%\bin\tpt_sampling.exe"
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%\src\sampling\sampling\oid2type.dat" "%PUBLISH_PATH%\lib\oid2type.dat"
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%\src\sampling\snmptools\snmptools.exe"  %PUBLISH_PATH%\tools\snmptools.exe
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%\src\lua_binding\lib\lua52_amd64.dll" "%PUBLISH_PATH%\bin\lua52_amd64.dll"
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%\src\lua_binding\lib\cjson_amd64.dll" "%PUBLISH_PATH%\bin\cjson_amd64.dll"
@if errorlevel 1 goto failed
xcopy /Y /S /E "%ENGINE_PATH%\src\lua_binding\microlight\*" %PUBLISH_PATH%\lib\microlight\
@if errorlevel 1 goto failed

:test_carrier
@if not defined is_test goto build_carrier
cd %ENGINE_PATH%\src\carrier
go test -v  %test_data_db_url_mysql% %test_delayed_job_db_url_mysql% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456
@if errorlevel 1 goto failed
go test -v  %test_data_db_url_mssql% %test_delayed_job_db_url_mssql% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456
@if errorlevel 1 goto failed
go test -v  %test_data_db_url% %test_delayed_job_db_url% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456
@if errorlevel 1 goto failed

:build_carrier
@if not defined is_compile goto install_carrier
cd %ENGINE_PATH%\src\carrier\carrier
del "*.exe"
go build
@if errorlevel 1 goto failed

:install_carrier
@if not defined is_install goto test_poller
copy "%ENGINE_PATH%\src\carrier\carrier\carrier.exe" "%PUBLISH_PATH%\bin\tpt_carrier.exe"
@if errorlevel 1 goto failed
set old_dir=%cd%
cd %ENGINE_PATH%\src\carrier\db
xcopy /Y /S /E /EXCLUDE:exclude_file "%ENGINE_PATH%\src\carrier\db"  %PUBLISH_PATH%\lib\data-migrations\
@if errorlevel 1 goto failed
cd %old_dir%




:test_poller
@if not defined is_test goto build_poller
cd %ENGINE_PATH%\src\poller
go test -v  %test_data_db_url_mysql% %test_delayed_job_db_url_mysql% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456
@if errorlevel 1 goto failed
go test -v  %test_data_db_url_mssql% %test_delayed_job_db_url_mssql% -db_table=tpt_delayed_jobs -not_limit=true -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456
@if errorlevel 1 goto failed
go test -v %test_db_url% %test_data_db_url% %test_delayed_job_db_url% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456
@if errorlevel 1 goto failed


go test -v  %test_data_db_url_mysql% %test_delayed_job_db_url_mysql% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_delayed_job_db_url_oracle_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed
go test -v  %test_data_db_url_mssql% %test_delayed_job_db_url_mssql% -db_table=tpt_delayed_jobs -not_limit=true -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_delayed_job_db_url_oracle_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed
go test -v %test_db_url% %test_data_db_url% %test_delayed_job_db_url% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_delayed_job_db_url_oracle_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed

go test -v  %test_data_db_url_mysql% %test_delayed_job_db_url_mysql% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_delayed_job_db_url_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed
go test -v  %test_data_db_url_mssql% %test_delayed_job_db_url_mssql% -db_table=tpt_delayed_jobs -not_limit=true -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456  %test_delayed_job_db_url_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed

go test -v  %test_data_db_url_mssql% %test_delayed_job_db_url_mssql% -db_table=tpt_delayed_jobs -not_limit=true -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_data_db_url_mysql_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed
go test -v %test_db_url% %test_data_db_url% %test_delayed_job_db_url% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_data_db_url_mysql_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed

go test -v  %test_data_db_url_mysql% %test_delayed_job_db_url_mysql% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_data_db_url_mssql_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed
go test -v %test_db_url% %test_data_db_url% %test_delayed_job_db_url% -db_table=tpt_delayed_jobs -redis=127.0.0.1:9456 -redis_address=127.0.0.1:9456 %test_data_db_url_mssql_for_test% -test.run=TestNotificationsForDb
@if errorlevel 1 goto failed


:build_poller
@if not defined is_compile goto install_poller
cd %ENGINE_PATH%\src\poller\poller
del "*.exe"
go build
@if errorlevel 1 goto failed
:install_poller
@if not defined is_install goto build_ok
copy "%ENGINE_PATH%\src\poller\poller\poller.exe" "%PUBLISH_PATH%\bin\tpt_poller.exe"
@if errorlevel 1 goto failed
xcopy /Y /S /E "%ENGINE_PATH%\src\poller\templates\*"   %PUBLISH_PATH%\lib\alerts\templates\
@if errorlevel 1 goto failed
copy "%ENGINE_PATH%\src\poller\poller_debug.txt" "%PUBLISH_PATH%\poller_debug.txt"
@if errorlevel 1 goto failed



copy "%ENGINE_PATH%\src\autostart_engine.conf" "%PUBLISH_PATH%\conf\autostart_engine.conf"
@if errorlevel 1 goto failed


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