# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM --platform=$BUILDPLATFORM alpine:3.21 AS certs
RUN apk add --no-cache ca-certificates

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/server ./server

USER 65532:65532

EXPOSE 8080

CMD ["./server"]
