FROM mariadb:10.5.2
COPY cmd/games/sql/data.sql /docker-entrypoint-initdb.d/