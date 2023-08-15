FROM golang:1.21 as builder

ENV GOOS=linux
WORKDIR /src
COPY . .

RUN go mod download
RUN make test
RUN make check
RUN make static

FROM gcr.io/distroless/static-debian11:nonroot
COPY --from=builder /src/bin/teams-backend /app/teams-backend
CMD ["/app/teams-backend"]
