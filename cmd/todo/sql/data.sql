DROP DATABASE IF EXISTS data;
CREATE DATABASE data;
USE data;

DROP TABLE IF EXISTS lists;
CREATE TABLE lists (
    user BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    name VARCHAR(100) NOT NULL,
    itemCount INT UNSIGNED NOT NULL,
    firstItem BINARY(16) NULL,
    PRIMARY KEY createdOn (user, createdOn, id),
    UNIQUE INDEX name (user, name, createdOn, id),
    UNIQUE INDEX id (user, id),
    UNIQUE INDEX itemCount (user, itemCount, createdOn, id)
);

DROP TABLE IF EXISTS items;
CREATE TABLE items (
    user BINARY(16) NOT NULL,
    list BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    name VARCHAR(100) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    nextItem BINARY(16) NULL,
    PRIMARY KEY createdOn (user, list, id)
);

DROP USER IF EXISTS 'data'@'%';
CREATE USER 'data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data.* TO 'data'@'%';
GRANT INSERT ON data.* TO 'data'@'%';
GRANT UPDATE ON data.* TO 'data'@'%';
GRANT DELETE ON data.* TO 'data'@'%';
GRANT EXECUTE ON data.* TO 'data'@'%';