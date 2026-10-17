FROM docker.io/library/golang:1.24 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/cloudlibrary .

FROM docker.io/bcliang/docker-libgourou:ubuntu

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/cloudlibrary /usr/local/bin/cloudlibrary

WORKDIR /home/libgourou

ENTRYPOINT ["/usr/local/bin/cloudlibrary"]
