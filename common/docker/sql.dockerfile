FROM mariadb:10.5.2
COPY common/sql/users.sql /docker-entrypoint-initdb.d/users.sql
COPY common/sql/pwds.sql /docker-entrypoint-initdb.d/pwds.sql
COPY common/sql/data.sql /docker-entrypoint-initdb.d/data.sql
COPY cmd/pandemic/sql/data.sql /docker-entrypoint-initdb.d/data_pandemic.sql
COPY cmd/todo/sql/data.sql /docker-entrypoint-initdb.d/data_todo.sql