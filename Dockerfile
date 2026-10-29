# -- Building stage ----------
FROM golang:1.27 AS builder

WORKDIR /workspace

COPY . .
RUN make build

# -- Production stage ----------
FROM debian:bookworm-slim AS production

WORKDIR /workspace

COPY --from=builder /workspace/build .

CMD ["/workspace/serenia"]

