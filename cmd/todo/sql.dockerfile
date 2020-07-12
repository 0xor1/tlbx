FROM mariadb:10.5.2
COPY cmd/todo/sql/users.sql /docker-entrypoint-initdb.d/
COPY cmd/todo/sql/pwds.sql /docker-entrypoint-initdb.d/
COPY cmd/todo/sql/data.sql /docker-entrypoint-initdb.d/