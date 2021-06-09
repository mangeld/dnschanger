FROM golang as builder

WORKDIR /build

COPY . /build/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build digitalocean_dyn_dns.go


FROM scratch

COPY --from=builder /build/digitalocean_dyn_dns .

CMD ["./digitalocean_dyn_dns"]