FROM --platform=linux/amd64 golang:1.23.2-bookworm
ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ENV AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID
ENV AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY

WORKDIR /go/src/url-shortener

COPY ./backend/url-shortener /usr/local/bin/url-shortener
RUN chmod +x /usr/local/bin/url-shortener

CMD ["/usr/local/bin/url-shortener"]