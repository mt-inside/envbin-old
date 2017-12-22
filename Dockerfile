FROM scratch

COPY envbin-linux /envbin
COPY main.html /

CMD ["/envbin"]
