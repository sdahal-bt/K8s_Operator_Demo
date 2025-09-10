# filepath: /Users/sdahal/Repos/K8s_Operator_Demo/secretswatcher-operator/Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o secretswatcher-operator main.go

FROM alpine:latest
COPY --from=builder /app/secretswatcher-operator /secretswatcher-operator
ENTRYPOINT ["/secretswatcher-operator"]