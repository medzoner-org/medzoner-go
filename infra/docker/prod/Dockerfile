#############################################
### Runtime image (binaires prébuild via CI/Makefile)
#############################################
FROM alpine:3.23.3 AS runner

# Un seul RUN = moins de layers, build et pull plus rapides.
RUN addgroup --gid 1000 -S app \
    && adduser --uid 1000 -S -G app app \
    && mkdir -p /opt

# Copie en tant qu'utilisateur app directement (évite un chown -R massif).
COPY --chown=app:app --chmod=0755 ./bin /opt/bin
COPY --chown=app:app ./tmpl /opt/tmpl
COPY --chown=app:app ./.var /opt/.var
COPY --chown=app:app ./public /opt/public
COPY --chown=app:app ./migrations /opt/migrations
COPY --chown=app:app ./internal/resources /opt/internal/resources

USER app
WORKDIR /opt

HEALTHCHECK --interval=5s --timeout=3s --retries=3 CMD ["lsof", "-i:8080", "||", "exit", "1"]

CMD ["/opt/bin/app"]

