FROM golang:alpine AS builder
RUN apk add --no-cache git

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build

FROM alpine:latest

RUN adduser -D app
USER app

COPY --from=builder /src/rmapi /usr/local/bin/rmapi
ENTRYPOINT ["rmapi"] 
