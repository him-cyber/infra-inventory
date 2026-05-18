FROM golang:1.26.3-alpine AS build
ARG SERVICE=api
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -buildid=" -o /out/service ./cmd/${SERVICE}

FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build --chown=65532:65532 /out/service /app/service
COPY --chown=65532:65532 config /app/config
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/service"]
