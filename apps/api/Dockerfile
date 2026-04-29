FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.work ./
COPY apps/api/go.mod apps/api/go.sum* ./apps/api/
WORKDIR /src/apps/api
ENV GOWORK=off
RUN go mod download

WORKDIR /src
COPY apps/api ./apps/api
WORKDIR /src/apps/api
ENV GOWORK=off
RUN go build -o /out/devsvault-api ./cmd/api

FROM alpine:3.20
RUN adduser -D -H devsvault
USER devsvault
COPY --from=build /out/devsvault-api /usr/local/bin/devsvault-api
EXPOSE 8080
ENTRYPOINT ["devsvault-api"]