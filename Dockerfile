FROM busybox:glibc

COPY ./gomoticasa /usr/bin/gomoticasa

EXPOSE 8080

CMD ["/usr/bin/gomoticasa"]