FROM mariadb:10.5.2
COPY sql/users.sql /docker-entrypoint-initdb.d/
COPY sql/pwds.sql /docker-entrypoint-initdb.d/
COPY cmd/games/sql/data.sql /docker-entrypoint-initdb.d/