FROM scratch
COPY ./rubik /rubik
ENTRYPOINT ["/rubik"]
