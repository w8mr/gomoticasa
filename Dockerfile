FROM debian:buster-slim

COPY ./main /usr/share/gomoticasa

EXPOSE 8080

CMD ["/usr/share/bin/gomoticasa"]