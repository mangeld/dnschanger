FROM golang as builder

WORKDIR /build

COPY . /build/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go


FROM scratch

COPY --from=builder /build/main .

CMD ["./main"]