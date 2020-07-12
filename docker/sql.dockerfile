FROM mariadb:10.5.2

# generic sqls for user pkg tests
COPY sql/users.sql /docker-entrypoint-initdb.d/users.sql
COPY sql/pwds.sql /docker-entrypoint-initdb.d/pwds.sql
COPY sql/data.sql /docker-entrypoint-initdb.d/data.sql

# games data sql
COPY cmd/games/sql/data.sql /docker-entrypoint-initdb.d/data_games.sql

# todo sqls
COPY cmd/todo/sql/users.sql /docker-entrypoint-initdb.d/users_todo.sql
COPY cmd/todo/sql/pwds.sql /docker-entrypoint-initdb.d/pwds_todo.sql
COPY cmd/todo/sql/data.sql /docker-entrypoint-initdb.d/data_todo.sql