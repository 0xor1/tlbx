DROP DATABASE IF EXISTS data_trees;
CREATE DATABASE data_trees;
USE data_trees;

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
    id BINARY(16) NOT NULL,
    alias VARCHAR(50) NOT NULL,
    hasAvatar BOOLEAN NOT NULL,
    createdOn DATETIME NOT NULL,
    isUser BOOLEAN NOT NULL,
    PRIMARY KEY id (id),
    UNIQUE INDEX alias (alias, id)
);

DROP TABLE IF EXISTS accountMembers;
CREATE TABLE accountMembers(
	account   BINARY(16) NOT NULL,
	user      BINARY(16) NOT NULL,
    PRIMARY KEY (account, user),
    UNIQUE INDEX (user, account)
);

DROP USER IF EXISTS 'data_trees'@'%';
CREATE USER 'data_trees'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_trees.* TO 'data_trees'@'%';
GRANT INSERT ON data_trees.* TO 'data_trees'@'%';
GRANT UPDATE ON data_trees.* TO 'data_trees'@'%';
GRANT DELETE ON data_trees.* TO 'data_trees'@'%';
GRANT EXECUTE ON data_trees.* TO 'data_trees'@'%';