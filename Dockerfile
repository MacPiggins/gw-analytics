FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /analytics ./cmd/main.go

FROM scratch
COPY --from=build /analytics /analytics
EXPOSE 8080
ENTRYPOINT ["/analytics"]