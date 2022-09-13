FROM golang:1.19.1-alpine as builder
RUN apk add --no-cache git make curl build-base
ENV GOOS=linux
COPY . /src
WORKDIR /src
RUN make test
RUN make vulncheck
RUN make alpine

FROM alpine:3.15
RUN apk add --no-cache ca-certificates tzdata
RUN export PATH=$PATH:/app
WORKDIR /app
COPY --from=builder /src/bin/console /app/console
CMD ["/app/console"]
