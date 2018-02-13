FROM alpine:3.7

RUN apk update && apk add bash && mkdir /app
COPY sese /app/

WORKDIR /app/
CMD ["./sese"]
