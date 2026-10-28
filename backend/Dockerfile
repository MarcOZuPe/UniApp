# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src

COPY backend/ ./backend/
COPY docs/ ./docs/

WORKDIR /src/backend
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /src/docs /app/docs

ENV PORT=8080
ENV SWAGGER_PORT=8081
EXPOSE 8080
EXPOSE 8081
CMD ["/app/server"]
