# docker build . -t helm-api -f helm-api.Dockerfile
# docker run helm-api
FROM golang:alpine AS builder

RUN mkdir /app
# ADD helm-chart /app/chart
ADD . /app/
WORKDIR /app
RUN go build -o helm-api .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app ./
CMD ["./helm-api"]
