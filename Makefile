
REPO = uluyol/startsync
VERSION = 0.1

.PHONY: all docker-build

all: pb/startsync.pb.go

pb/startsync.pb.go: pb/startsync.proto
	protoc --go_out=plugins=grpc:. pb/startsync.proto

startsyncd startsync:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build github.com/uluyol/startsync/cmd/$@

docker-build: pb/startsync.pb.go startsyncd startsync
	docker build -t $(REPO):$(VERSION) .

docker-push: docker-build
	docker push $(REPO):$(VERSION)

clean:
	rm -f startsync startsyncd
