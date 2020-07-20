FROM mariadb:10.5.2

# generic sqls for user pkg tests
COPY sql/users.sql /docker-entrypoint-initdb.d/users.sql
COPY sql/pwds.sql /docker-entrypoint-initdb.d/pwds.sql
COPY sql/data.sql /docker-entrypoint-initdb.d/data.sql

# games data sql
COPY cmd/games/sql/data.sql /docker-entrypoint-initdb.d/games_data.sql

# todo sqls
COPY cmd/todo/sql/users.sql /docker-entrypoint-initdb.d/todo_users.sql
COPY cmd/todo/sql/pwds.sql /docker-entrypoint-initdb.d/todo_pwds.sql
COPY cmd/todo/sql/data.sql /docker-entrypoint-initdb.d/todo_data.sql

# trees sqls
COPY cmd/trees/sql/users.sql /docker-entrypoint-initdb.d/trees_users.sql
COPY cmd/trees/sql/pwds.sql /docker-entrypoint-initdb.d/trees_pwds.sql
COPY cmd/trees/sql/data.sql /docker-entrypoint-initdb.d/trees_data.sql