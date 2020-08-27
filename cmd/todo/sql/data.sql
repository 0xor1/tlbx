SET GLOBAL max_recursive_iterations=2000;
DROP DATABASE IF EXISTS todo_data;
CREATE DATABASE todo_data
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE todo_data;

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

DROP USER IF EXISTS 'todo_data'@'%';
CREATE USER 'todo_data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON todo_data.* TO 'todo_data'@'%';
GRANT INSERT ON todo_data.* TO 'todo_data'@'%';
GRANT UPDATE ON todo_data.* TO 'todo_data'@'%';
GRANT DELETE ON todo_data.* TO 'todo_data'@'%';
GRANT EXECUTE ON todo_data.* TO 'todo_data'@'%';