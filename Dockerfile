FROM scratch
COPY ./build/rubik /rubik
ENTRYPOINT ["/rubik"]

