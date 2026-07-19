# syntax=docker/dockerfile:1

# ---- Build stage ----------------------------------------------------------
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO disabled (modernc.org/sqlite is pure Go) and cross-compiled for the
# Lambda Graviton (arm64) runtime. The binary MUST be named "bootstrap" so
# the provided.al2023 base image picks it up as the Lambda entrypoint.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bootstrap .

# ---- Final Lambda image ---------------------------------------------------
FROM public.ecr.aws/lambda/provided:al2023

COPY --from=builder /app/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap

ENTRYPOINT [ "/var/runtime/bootstrap" ]
