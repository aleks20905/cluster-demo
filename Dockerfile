FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev git

WORKDIR /build

RUN go install github.com/a-h/templ/cmd/templ@latest

ENV PATH="$PATH:/go/bin"

COPY . .

RUN go mod download

RUN templ generate

RUN go build -o ./bin/userapi ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /build/bin/userapi ./userapi

VOLUME ["/app/db"]

EXPOSE 8080

CMD ["/app/userapi"]

