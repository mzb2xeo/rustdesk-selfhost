# ==============================================================================
# STAGE 1: Build Vue 3 Web Admin Frontend
# ==============================================================================
FROM node:18-alpine AS frontend-builder
WORKDIR /frontend

# Copy dependencies list first
COPY ./rustdesk-api-web/package.json ./rustdesk-api-web/package-lock.json* ./
RUN npm install

# Copy source and build
COPY ./rustdesk-api-web/ ./
RUN npm run build

# ==============================================================================
# STAGE 2: Build Go Backend
# ==============================================================================
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app

# Install dependencies required for CGO compilation (needed for sqlite3)
RUN apk add --no-cache gcc musl-dev git

# Install swag tool for generating swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy dependency configuration
COPY ./rustdesk-api/go.mod ./rustdesk-api/go.sum ./
RUN go mod download

# Copy source code
COPY ./rustdesk-api/ ./

# Generate Swagger Docs
RUN swag init -g cmd/apimain.go --output docs/api --instanceName api --exclude http/controller/admin && \
    swag init -g cmd/apimain.go --output docs/admin --instanceName admin --exclude http/controller/api

# Build backend binary with CGO enabled (for sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-s -w -extldflags '-static'" -o release/apimain cmd/apimain.go

# ==============================================================================
# STAGE 3: Final Production Image
# ==============================================================================
FROM alpine:latest
WORKDIR /app

# Install timezone data
RUN apk add --no-cache tzdata ca-certificates

# Copy build artifacts and configuration folders
COPY --from=backend-builder /app/release/apimain /app/apimain
COPY --from=backend-builder /app/conf /app/conf/
COPY --from=backend-builder /app/resources /app/resources/
COPY --from=backend-builder /app/docs /app/docs/

# Copy compiled Web Admin frontend from Stage 1 to resources/admin
COPY --from=frontend-builder /frontend/dist/ /app/resources/admin/

# Create necessary runtime folders
RUN mkdir -p /app/data && mkdir -p /app/runtime

VOLUME /app/data
EXPOSE 21114

CMD ["./apimain"]
