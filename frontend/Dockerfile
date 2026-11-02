# syntax=docker/dockerfile:1

# --- Etapa 1: compilar la SPA ---
FROM node:24-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vite incorpora VITE_API_URL al compilar, no al ejecutar: hay que pasarla como argumento.
ARG VITE_API_URL=http://localhost:8080/api/v1
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# --- Etapa 2: Nginx sin root sirviendo dist/ ---
FROM nginxinc/nginx-unprivileged:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
