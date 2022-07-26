#!/bin/sh
#Copyright (C) 2020 iyfiysi
#gen by iyfiysi at 2021 Jun 17

# pb构建
cd proto
sh gen.sh
cd -

statik -src ./swagger
statik -src=internal/pkg/html -dest=internal/pkg -p=data

BUILT_AT=$(date "+%Y-%m-%d %H:%M:%S") #2020-03-30 10:08:54
COMMIT_TAG=$(git rev-parse HEAD) #1c7caa847ce196f0668e01794d3cd773944f3127
if [ ${#COMMIT_TAG} -eq 40 ];then
    COMMIT_TAG=${COMMIT_TAG:0:8}
else
    COMMIT_TAG="unknow"
fi

eval $(go env)

# build gateway
cd cmd/gateway
go build -ldflags "-X 'main.commit=$COMMIT_TAG' -X 'main.date=$BUILT_AT'" -o short_url_gateway .
#GOARCH=amd64 GOOS=darwin go build -ldflags "-X 'main.commit=$COMMIT_TAG' -X 'main.date=$BUILT_AT'" -o short_url_gateway .
#GOARCH=amd64 GOOS=windows go build -ldflags "-X 'main.commit=$COMMIT_TAG' -X 'main.date=$BUILT_AT'" -o short_url_gateway .
mv short_url_gateway ../..
cd -

cd cmd/server
go build -ldflags "-X 'main.commit=$COMMIT_TAG' -X 'main.date=$BUILT_AT'" -o short_url_server .
mv short_url_server ../..
cd -

cd cmd/conf
go build -ldflags "-X 'main.commit=$COMMIT_TAG' -X 'main.date=$BUILT_AT'" -o short_url_conf .
mv short_url_conf ../..
cd -

#end
