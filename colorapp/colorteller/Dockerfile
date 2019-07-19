FROM golang:1.12 AS builder

WORKDIR /src
# cache layers for faster docker builds
COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o app .

FROM scratch
COPY --from=builder /src/app ./
ENTRYPOINT ["./app"]
