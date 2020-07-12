DROP DATABASE IF EXISTS users_tree;
CREATE DATABASE users_tree;
USE users_tree;

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

DROP USER IF EXISTS 'users_tree'@'%';
CREATE USER 'users_tree'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON users_tree.* TO 'users_tree'@'%';
GRANT INSERT ON users_tree.* TO 'users_tree'@'%';
GRANT UPDATE ON users_tree.* TO 'users_tree'@'%';
GRANT DELETE ON users_tree.* TO 'users_tree'@'%';
GRANT EXECUTE ON users_tree.* TO 'users_tree'@'%';