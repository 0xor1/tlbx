DROP DATABASE IF EXISTS users_todo;
CREATE DATABASE users_todo;
USE users_todo;

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

DROP USER IF EXISTS 'users_todo'@'%';
CREATE USER 'users_todo'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON users_todo.* TO 'users_todo'@'%';
GRANT INSERT ON users_todo.* TO 'users_todo'@'%';
GRANT UPDATE ON users_todo.* TO 'users_todo'@'%';
GRANT DELETE ON users_todo.* TO 'users_todo'@'%';
GRANT EXECUTE ON users_todo.* TO 'users_todo'@'%';