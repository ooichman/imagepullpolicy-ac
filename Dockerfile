FROM golang:alpine AS builder

RUN mkdir -p /opt/app-root/src/ippac
WORKDIR /opt/app-root/src/ippac
ENV GOPATH=/opt/app-root/
ENV PATH="${PATH}:/opt/app-root/src/go/bin/"
COPY  src/ippac .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ippac

FROM scratch
COPY --from=builder  /etc/passwd /etc/passwd
COPY --from=builder  /opt/app-root/src/ippac/ippac /usr/bin/
USER nobody

EXPOSE 8080 8443

CMD ["/usr/bin/ippac"]