FROM debian:buster-slim

COPY ./gomoticasa /usr/bin/gomoticasa

EXPOSE 8080

CMD ["/usr/bin/gomoticasa"]