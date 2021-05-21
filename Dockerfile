# docker build . -t helm-api
# docker run helm-api
FROM golang:alpine

RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN go build -o helm-api .

CMD ["./helm-api"]
