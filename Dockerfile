FROM debian:buster-slim

COPY ./gomoticasa /usr/share/gomoticasa

EXPOSE 8080

CMD ["/usr/share/bin/gomoticasa"]