FROM cgr.dev/chainguard/go:latest as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO=0 go build -o /app/bitbom main.go

FROM cgr.dev/chainguard/go:latest
WORKDIR /app
COPY --from=builder /app/bitbom /app/bitbom

ENTRYPOINT ["/app/bitbom"]
