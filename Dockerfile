FROM alpine:3.3

RUN apk add --update bash ca-certificates curl
RUN mkdir -p /opt/bin && \
		curl -Lo /opt/bin/s3kms https://s3-us-west-2.amazonaws.com/opsee-releases/go/vinz-clortho/s3kms-linux-amd64 && \
    chmod 755 /opt/bin/s3kms && \
    curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate

ENV BEZOSPHERE_ADDRESS ""
ENV BEZOSPHERE_SPANX_ADDRESS ""
ENV BEZOSPHERE_CERT "cert.pem"
ENV BEZOSPHERE_CERT_KEY "key.pem"
ENV BEZOSPHERE_POSTGRES_CONN ""
ENV APPENV ""

COPY run.sh /
COPY key.pem /
COPY cert.pem /
COPY migrations /migrations
COPY target/linux/amd64/bin/* /

EXPOSE 9104
CMD ["/bezosphere"]
