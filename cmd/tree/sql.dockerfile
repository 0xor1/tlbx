FROM mariadb:10.5.2
COPY cmd/tree/sql/users.sql /docker-entrypoint-initdb.d/
COPY cmd/tree/sql/pwds.sql /docker-entrypoint-initdb.d/
COPY cmd/tree/sql/data.sql /docker-entrypoint-initdb.d/