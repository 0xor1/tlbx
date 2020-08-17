DROP DATABASE IF EXISTS todo_users;
CREATE DATABASE todo_users
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE todo_users;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id BINARY(16) NOT NULL,
	email VARCHAR(250) NOT NULL,
    handle VARCHAR(15) NULL,
    alias VARCHAR(50) NULL,
    hasAvatar BOOLEAN NULL,
    registeredOn DATETIME(3) NOT NULL,
    activatedOn DATETIME(3) NULL,
	newEmail VARCHAR(250) NULL,
	activateCode VARCHAR(250) NULL,
	changeEmailCode VARCHAR(250) NULL,
	lastPwdResetOn DATETIME(3) NULL,
    PRIMARY KEY email (email),
    UNIQUE INDEX id (id),
    UNIQUE INDEX handle (handle)
);

DROP USER IF EXISTS 'todo_users'@'%';
CREATE USER 'todo_users'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON todo_users.* TO 'todo_users'@'%';
GRANT INSERT ON todo_users.* TO 'todo_users'@'%';
GRANT UPDATE ON todo_users.* TO 'todo_users'@'%';
GRANT DELETE ON todo_users.* TO 'todo_users'@'%';
GRANT EXECUTE ON todo_users.* TO 'todo_users'@'%';