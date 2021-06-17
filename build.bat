@echo on

@rem Copyright (C) 2020 iyfiysi
@rem gen by iyfiysi at 2021 Jun 16

@rem pb构建
cd proto
call gen.bat
cd ..

%GOPATH%\bin\statik -src=swagger -f
%GOPATH%\bin\statik -src=internal\pkg\html -dest=internal\pkg -p=data -f

rem BUILT_AT=2020-03-30 10:08:54
set BUILT_AT=%date:~0,4%-%date:~5,2%-%date:~8,2% %time%
set COMMIT_TAG=unknow
for /F %%i in ('git rev-parse HEAD') do ( set COMMIT_TAG=%%i)

cd cmd\gateway
go build -ldflags "-X 'main.commit=%COMMIT_TAG%' -X 'main.date=%BUILT_AT%'" -o short_url_gateway.exe .
move short_url_gateway.exe ..\..
cd ..\..

cd cmd\server
go build -ldflags "-X 'main.commit=%COMMIT_TAG%' -X 'main.date=%BUILT_AT%'" -o short_url_server.exe .
move short_url_server.exe ..\..
cd ..\..

cd cmd\conf
go build -ldflags "-X 'main.commit=%COMMIT_TAG%' -X 'main.date=%BUILT_AT%'" -o short_url_conf.exe .
move short_url_conf.exe ..\..
cd ..\..