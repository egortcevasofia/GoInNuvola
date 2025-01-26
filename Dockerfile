FROM scratch
COPY kvs .
EXPOSE 8080
CMD ["/kvs"]

