FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.work ./
COPY apps/cli/go.mod apps/cli/go.sum* ./apps/cli/
WORKDIR /src/apps/cli
ENV GOWORK=off
RUN go mod download

WORKDIR /src
COPY apps/cli ./apps/cli
WORKDIR /src/apps/cli
ENV GOWORK=off
RUN go build -o /out/devsvault ./cmd/devsvault

FROM alpine:latest
RUN adduser -D -H devsvault
USER devsvault
COPY --from=build /out/devsvault /usr/local/bin/devsvault
ENTRYPOINT ["devsvault"]