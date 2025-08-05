# -------- Build stage --------
FROM golang:1.23-alpine AS builder

ARG APP_PORT=4000

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    PROJECT_DIR=/app

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./src ./src
RUN go build -o main ./src/main.go

# -------- Final stage --------
FROM gcr.io/distroless/static:nonroot

ARG APP_PORT
ENV APP_PORT=${APP_PORT}
EXPOSE ${APP_PORT}

WORKDIR /app

COPY --from=builder /app/main .

USER nonroot:nonroot

ENTRYPOINT ["/app/main", "-routes=true"]
