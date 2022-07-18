FROM golang as builder
ADD . /src
WORKDIR /src

RUN CGO_ENABLED=0 go build -o app .

FROM scratch
COPY --from=builder /src/app /
CMD ["/app"]
