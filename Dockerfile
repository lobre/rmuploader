FROM golang:alpine as builder
COPY . /app
WORKDIR /app
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -o rmuploader

FROM alpine:latest
RUN apk add --no-cache wkhtmltopdf
WORKDIR /root/
COPY --from=builder /app .
EXPOSE 8080
ENTRYPOINT ["./rmuploader"]
