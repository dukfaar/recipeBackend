FROM golang as builder
ARG DB_HOST
ARG PORT
ARG NSQD_TCP_URL
ARG NSQLOOKUP_HTTP_URL

ENV DB_HOST=$DB_HOST
ENV PORT=$PORT
ENV NSQD_TCP_URL=$NSQD_TCP_URL
ENV NSQLOOKUP_HTTP_URL=$NSQLOOKUP_HTTP_URL

ENV GOBIN /go/bin

EXPOSE $PORT

RUN mkdir /app
RUN mkdir /go/src/app
COPY . /go/src/app
WORKDIR /go/src/app

RUN go get -u github.com/golang/dep/...
RUN dep ensure

RUN go build -o /app/main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

CMD ["/app/main"]