FROM cgr.dev/chainguard/go:latest as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /app/minefield main.go

FROM cgr.dev/chainguard/glibc-dynamic
WORKDIR /app
COPY --from=builder /app/minefield /app/minefield

ENTRYPOINT ["/app/minefield"]
