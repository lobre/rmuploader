FROM golang:alpine as builder
COPY . /app
WORKDIR /app
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -o rmuploader

FROM openlabs/docker-wkhtmltopdf:latest

ENV XDG_RUNTIME_DIR=/run/user/abc
EXPOSE 8080

# Set abc user
RUN useradd -ms /bin/bash abc && \
    mkdir -p /run/user/abc && chown abc /run/user/abc
USER abc
WORKDIR /home/abc

# Copy application
COPY --from=builder --chown=abc:abc /app/rmuploader .
COPY --from=builder --chown=abc:abc /app/web ./web

ENTRYPOINT ["./rmuploader"]
