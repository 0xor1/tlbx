FROM mariadb:10.5.2
COPY cmd/trees/sql/users.sql /docker-entrypoint-initdb.d/
COPY cmd/trees/sql/pwds.sql /docker-entrypoint-initdb.d/
COPY cmd/trees/sql/data.sql /docker-entrypoint-initdb.d/