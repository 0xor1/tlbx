FROM mariadb:10.5.2
COPY sql/users.sql /docker-entrypoint-initdb.d/users.sql
COPY sql/pwds.sql /docker-entrypoint-initdb.d/pwds.sql
COPY sql/data.sql /docker-entrypoint-initdb.d/data.sql
COPY cmd/games/sql/data.sql /docker-entrypoint-initdb.d/data_games.sql
COPY cmd/todo/sql/data.sql /docker-entrypoint-initdb.d/data_todo.sql