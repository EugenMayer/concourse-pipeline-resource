FROM golang:alpine as builder
RUN mkdir -p /assets
COPY bin/fly /assets/fly
COPY . /go/src/github.com/eugenmayer/concourse-pipeline-resource
ENV CGO_ENABLED 0
RUN go build -o /assets/in github.com/eugenmayer/concourse-pipeline-resource/cmd/in
RUN go build -o /assets/out github.com/eugenmayer/concourse-pipeline-resource/cmd/out
RUN go build -o /assets/check github.com/eugenmayer/concourse-pipeline-resource/cmd/check
RUN set -e; for pkg in $(go list ./... | grep -v "acceptance"); do \
		go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM alpine:edge AS resource
RUN apk add --no-cache bash tzdata ca-certificates curl
COPY --from=builder assets/ /opt/resource/
COPY bin/oauth_login.sh /opt/resource/oauth_login
RUN chmod +x /opt/resource/*

FROM resource AS tests
COPY --from=builder /tests /go-tests
RUN set -e; for test in /go-tests/*.test; do \
		$test; \
	done

FROM resource
