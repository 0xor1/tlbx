FROM mariadb:10.5.2
COPY cmd/pandemic/sql/data.sql /docker-entrypoint-initdb.d/