#############################################
### Build the Go Binary
#############################################
FROM golang:1.26.4 as build

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

RUN useradd -u 1000 medzoner
#RUN addgroup --gid 1000 -S app && adduser --uid 1000 -S -G app app

#############################################
### Run the Go Binary
#############################################
FROM alpine:3.23.3 AS runner

COPY ./bin /opt/bin
COPY ./tmpl /opt/tmpl
COPY ./.var /opt/.var
COPY ./public /opt/public
COPY ./migrations /opt/migrations
COPY ./internal/resources /opt/internal/resources

#COPY --from=build /etc/passwd /etc/passwd
#COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
RUN addgroup --gid 1000 -S app && adduser --uid 1000 -S -G app app

RUN chown -R app:app /opt

USER app

WORKDIR /opt

HEALTHCHECK --interval=5s --timeout=3s --retries=3 CMD ["lsof", "-i:8080", "||", "exit", "1"]

CMD ["/opt/bin/app"]
