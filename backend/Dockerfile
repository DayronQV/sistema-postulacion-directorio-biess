# syntax=docker/dockerfile:1

# --- Etapa 1: compilar un binario estático ---
FROM golang:1.26-alpine AS build
WORKDIR /src

# Primero las dependencias, para aprovechar la caché de capas.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# --- Etapa 2: imagen mínima, sin shell y sin root ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
