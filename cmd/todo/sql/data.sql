DROP DATABASE IF EXISTS data_todo;
CREATE DATABASE data_todo;
USE data_todo;

DROP TABLE IF EXISTS lists;
CREATE TABLE lists (
    user BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    name VARCHAR(250) NOT NULL,
    todoItemCount INT UNSIGNED NOT NULL,
    completedItemCount INT UNSIGNED NOT NULL,
    PRIMARY KEY createdOn (user, createdOn),
    UNIQUE INDEX name (user, name, createdOn),
    UNIQUE INDEX id (user, id),
    UNIQUE INDEX todoItemCount (user, todoItemCount, createdOn),
    UNIQUE INDEX completedItemCount (user, completedItemCount, createdOn)
);

DROP TABLE IF EXISTS items;
CREATE TABLE items (
    user BINARY(16) NOT NULL,
    list BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    name VARCHAR(250) NOT NULL,
    completedOn DATETIME(3) NOT NULL, -- not null, use go zero time for null
    PRIMARY KEY createdOn (user, list, completedOn, createdOn),
    UNIQUE INDEX name (user, list, completedOn, name, createdOn),
    UNIQUE INDEX id (user, list, id)
);

DROP USER IF EXISTS 'data_todo'@'%';
CREATE USER 'data_todo'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_todo.* TO 'data_todo'@'%';
GRANT INSERT ON data_todo.* TO 'data_todo'@'%';
GRANT UPDATE ON data_todo.* TO 'data_todo'@'%';
GRANT DELETE ON data_todo.* TO 'data_todo'@'%';
GRANT EXECUTE ON data_todo.* TO 'data_todo'@'%';