FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY main.go main.go
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /bin/decentrss

FROM scratch AS runtime
COPY --from=builder /bin/decentrss /bin/decentrss
WORKDIR /app
CMD [ "decentrss" ]
