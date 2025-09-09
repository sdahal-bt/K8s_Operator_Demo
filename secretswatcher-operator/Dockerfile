FROM golang:1.25-alpine as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o secretswatcher-operator main.go

FROM gcr.io/distroless/base
COPY --from=builder /app/secretswatcher-operator /secretswatcher-operator
ENTRYPOINT ["/secretswatcher-operator"]