DROP DATABASE IF EXISTS data_tree;
CREATE DATABASE data_tree;
USE data_tree;

DROP TABLE IF EXISTS lists;
CREATE TABLE lists (
    user BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    name VARCHAR(250) NOT NULL,
    treeItemCount INT UNSIGNED NOT NULL,
    completedItemCount INT UNSIGNED NOT NULL,
    PRIMARY KEY createdOn (user, createdOn),
    UNIQUE INDEX name (user, name, createdOn),
    UNIQUE INDEX id (user, id),
    UNIQUE INDEX treeItemCount (user, treeItemCount, createdOn),
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

DROP USER IF EXISTS 'data_tree'@'%';
CREATE USER 'data_tree'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_tree.* TO 'data_tree'@'%';
GRANT INSERT ON data_tree.* TO 'data_tree'@'%';
GRANT UPDATE ON data_tree.* TO 'data_tree'@'%';
GRANT DELETE ON data_tree.* TO 'data_tree'@'%';
GRANT EXECUTE ON data_tree.* TO 'data_tree'@'%';