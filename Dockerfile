# Builder
FROM golang:bullseye AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o hyperdot-node cmd/node/main.go

# Runtime
FROM alpine:3.18.4

WORKDIR /app

COPY --from=builder /app/hyperdot-node .

EXPOSE 3030

ENV HYPETDOT_NODE_CONFIG /app/config/hyperdot.json
ENV GOOGLE_APPLICATION_CREDENTIALS=/app/config/hyperdot-gcloud-iam.json

COPY ./config/hyperdot.json /app/config/hyperdot.json
COPY ./config/hyperdot-gcloud-iam.json /app/config/hyperdot-gcloud-iam.json

CMD ["./hyperdot-node", "-config=/app/config/hyperdot.json"]
