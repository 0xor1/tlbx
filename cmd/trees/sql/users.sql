DROP DATABASE IF EXISTS users_trees;
CREATE DATABASE users_trees;
USE users_trees;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id BINARY(16) NOT NULL,
	email VARCHAR(250) NOT NULL,
    alias VARCHAR(50) NULL,
    hasAvatar BOOLEAN NULL,
    registeredOn DATETIME NOT NULL,
    activatedOn DATETIME NULL,
	newEmail VARCHAR(250) NULL,
	activateCode VARCHAR(250) NULL,
	changeEmailCode VARCHAR(250) NULL,
	lastPwdResetOn DATETIME NULL,
    PRIMARY KEY email (email),
    UNIQUE INDEX id (id)
);

DROP USER IF EXISTS 'users_trees'@'%';
CREATE USER 'users_trees'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON users_trees.* TO 'users_trees'@'%';
GRANT INSERT ON users_trees.* TO 'users_trees'@'%';
GRANT UPDATE ON users_trees.* TO 'users_trees'@'%';
GRANT DELETE ON users_trees.* TO 'users_trees'@'%';
GRANT EXECUTE ON users_trees.* TO 'users_trees'@'%';