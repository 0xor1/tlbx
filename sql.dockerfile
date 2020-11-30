FROM mariadb:10.5.2 AS builder

# That file does the DB initialization but also runs mysql daemon, by removing the last line it will only init
RUN ["sed", "-i", "s/exec \"$@\"/echo \"not running $@\"/", "/usr/local/bin/docker-entrypoint.sh"]

ENV MYSQL_ROOT_PASSWORD=dev

COPY sql/users.sql /docker-entrypoint-initdb.d/
COPY sql/pwds.sql /docker-entrypoint-initdb.d/
COPY sql/data.sql /docker-entrypoint-initdb.d/

RUN ["/usr/local/bin/docker-entrypoint.sh", "mysqld", "--datadir", "/initialized-db", "--aria-log-dir-path", "/initialized-db"]

FROM mariadb:10.5.2

COPY --from=builder /initialized-db /var/lib/mysql