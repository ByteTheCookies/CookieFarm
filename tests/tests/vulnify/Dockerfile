FROM ubuntu:latest

RUN apt update && apt install -y socat build-essential

WORKDIR /app
COPY vulnify.c vulnify.h justfile ./

RUN mkdir data
RUN apk add --no-cache just
RUN just build
RUN chmod +x ./vulnify
RUN rm -f ./vulnify.c ./vulnify.h ./justfile

EXPOSE 1337

CMD ["socat", "-t", "10", "-T", "10", "TCP-LISTEN:1337,reuseaddr,fork", "EXEC:/app/vulnify"]
