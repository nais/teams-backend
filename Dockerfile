FROM cgr.dev/chainguard/go:1.20 as builder

ENV GOOS=linux
WORKDIR /src
COPY . .

RUN go mod download
RUN make test
RUN make check
RUN make static

FROM cgr.dev/chainguard/static
COPY --from=builder /src/bin/teams-backend /app/teams-backend
CMD ["/app/teams-backend"]
