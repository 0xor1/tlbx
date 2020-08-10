DROP DATABASE IF EXISTS users;
CREATE DATABASE users
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_general_ci';
USE users;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id BINARY(16) NOT NULL,
	email VARCHAR(250) NOT NULL,
    handle VARCHAR(15) NULL,
    alias VARCHAR(50) NULL,
    hasAvatar BOOLEAN NULL,
    registeredOn DATETIME NOT NULL,
    activatedOn DATETIME NULL,
	newEmail VARCHAR(250) NULL,
	activateCode VARCHAR(250) NULL,
	changeEmailCode VARCHAR(250) NULL,
	lastPwdResetOn DATETIME NULL,
    PRIMARY KEY email (email),
    UNIQUE INDEX id (id),
    UNIQUE INDEX handle (handle)
);

DROP USER IF EXISTS 'users'@'%';
CREATE USER 'users'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON users.* TO 'users'@'%';
GRANT INSERT ON users.* TO 'users'@'%';
GRANT UPDATE ON users.* TO 'users'@'%';
GRANT DELETE ON users.* TO 'users'@'%';
GRANT EXECUTE ON users.* TO 'users'@'%';