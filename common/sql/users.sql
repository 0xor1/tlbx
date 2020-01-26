DROP DATABASE IF EXISTS users;
CREATE DATABASE users;
USE users;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id BINARY(16) NOT NULL,
	email VARCHAR(250) NOT NULL,
    name VARCHAR(100) NOT NULL,
    registeredOn DATETIME NOT NULL,
    activatedOn DATETIME NULL,
	newEmail VARCHAR(250) NULL,
	activateCode VARCHAR(250) NULL,
	confirmChangeEmailCode VARCHAR(250) NULL,
	resetPwdCode VARCHAR(250) NULL,
    PRIMARY KEY name (email),
    UNIQUE INDEX id (id)
);

DROP USER IF EXISTS 'users'@'%';
CREATE USER 'users'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON users.* TO 'users'@'%';
GRANT INSERT ON users.* TO 'users'@'%';
GRANT UPDATE ON users.* TO 'users'@'%';
GRANT DELETE ON users.* TO 'users'@'%';
GRANT EXECUTE ON users.* TO 'users'@'%';