FROM golang:1.13 as builder

ENV GO111MODULE=off
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/contextcloud/ccb-cli
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN go test $(go list ./... | grep -v /vendor/ | grep -v /template/|grep -v /build/|grep -v /sample/) -cover

RUN VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT=$(git rev-list -1 HEAD) \
    && go build --ldflags "-s -w \
    -X github.com/contextcloud/ccb-cli/version.GitCommit=${GIT_COMMIT} \
    -X github.com/contextcloud/ccb-cli/version.Version=${VERSION} \
    -X github.com/contextcloud/ccb-cli/commands.Platform=x86_64" \
    -a -installsuffix cgo -o ccb cli/main.go

FROM gcr.io/cloud-builders/docker

ENV PATH=$PATH:/usr/bin/
CMD ["ccb"]

COPY --from=builder /go/src/github.com/contextcloud/ccb-cli/ccb /usr/bin/