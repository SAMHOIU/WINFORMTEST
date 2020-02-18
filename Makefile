
GO_CMD=go
REPO_PATH=crawler.club/crawler
GIT_SHA=`git rev-parse --short HEAD || echo "GitNotFound"`
GO_LDFLAGS=-ldflags "-X ${REPO_PATH}/version.GitSHA=${GIT_SHA}"