DROP DATABASE IF EXISTS data_trees;
CREATE DATABASE data_trees;
USE data_trees;

DROP TABLE IF EXISTS lists;
CREATE TABLE lists (
    user BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    name VARCHAR(250) NOT NULL,
    treesItemCount INT UNSIGNED NOT NULL,
    completedItemCount INT UNSIGNED NOT NULL,
    PRIMARY KEY createdOn (user, createdOn),
    UNIQUE INDEX name (user, name, createdOn),
    UNIQUE INDEX id (user, id),
    UNIQUE INDEX treesItemCount (user, treesItemCount, createdOn),
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

DROP USER IF EXISTS 'data_trees'@'%';
CREATE USER 'data_trees'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_trees.* TO 'data_trees'@'%';
GRANT INSERT ON data_trees.* TO 'data_trees'@'%';
GRANT UPDATE ON data_trees.* TO 'data_trees'@'%';
GRANT DELETE ON data_trees.* TO 'data_trees'@'%';
GRANT EXECUTE ON data_trees.* TO 'data_trees'@'%';