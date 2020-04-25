FROM mariadb:10.5.2
COPY common/sql/users.sql /docker-entrypoint-initdb.d/
COPY common/sql/pwds.sql /docker-entrypoint-initdb.d/
COPY cmd/pandemic/sql/data.sql /docker-entrypoint-initdb.d/