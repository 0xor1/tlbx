DROP DATABASE IF EXISTS data;
CREATE DATABASE data;
USE data;

DROP TABLE IF EXISTS lists;
CREATE TABLE lists (
    user BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    name VARCHAR(100) NOT NULL,
    todoItemCount INT UNSIGNED NOT NULL,
    completedItemCount INT UNSIGNED NOT NULL,
    firstItem BINARY(16) NULL,
    PRIMARY KEY createdOn (user, createdOn, id),
    UNIQUE INDEX name (user, name, createdOn, id),
    UNIQUE INDEX id (user, id),
    UNIQUE INDEX todoItemCount (user, todoItemCount, createdOn, id),
    UNIQUE INDEX completedItemCount (user, completedItemCount, createdOn, id)
);

DROP TABLE IF EXISTS items;
CREATE TABLE items (
    user BINARY(16) NOT NULL,
    list BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    name VARCHAR(100) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    completedOn DATETIME(3) NOT NULL, -- not null, use go zero time for null
    nextItem BINARY(16) NULL,
    PRIMARY KEY id (user, list, completedOn, id)
);

DROP USER IF EXISTS 'data'@'%';
CREATE USER 'data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data.* TO 'data'@'%';
GRANT INSERT ON data.* TO 'data'@'%';
GRANT UPDATE ON data.* TO 'data'@'%';
GRANT DELETE ON data.* TO 'data'@'%';
GRANT EXECUTE ON data.* TO 'data'@'%';