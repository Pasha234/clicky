FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /collector ./cmd/collector

FROM alpine:3.21

COPY --from=build /collector /collector

EXPOSE 3000

ENTRYPOINT ["/collector"]
