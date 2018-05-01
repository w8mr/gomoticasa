FROM kpeu3i/go-arm-sdl2:1.9 as builder

ENV GOPATH=/go REPO_NAME=w8mr.nl PROJECT_NAME=go_my_home

COPY . $GOPATH/src/$REPO_NAME/$PROJECT_NAME
WORKDIR $GOPATH/src/$REPO_NAME/$PROJECT_NAME

RUN env GOARCH=arm GOARM=6 go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN env GOARCH=arm GOARM=6 go build -ldflags "-linkmode external -extldflags -static" -a -installsuffix cgo -o main .

RUN cp $GOPATH/src/$REPO_NAME/$PROJECT_NAME/main /go

FROM scratch

COPY --from=builder /go/main /

EXPOSE 8080

CMD ["/main"]