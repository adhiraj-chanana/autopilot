# STEP 1: Build stage
FROM golang:1.25 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o kube-autopilot .

# STEP 2: Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/kube-autopilot /app/kube-autopilot

# (Optional) Non-root user
RUN adduser -D autopilot
USER autopilot

ENTRYPOINT ["/app/kube-autopilot"]
