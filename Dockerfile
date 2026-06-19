FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum  ./
RUN go mod download

COPY . .

RUN apk add --no-cache gcc musl-dev

RUN CGO_ENABLED=1 GOOS=linux go build -tags no_gui -o /eko .

FROM alpine:latest AS build-release

WORKDIR /

COPY --from=build /eko /eko

EXPOSE 7700

CMD ["/eko"]

