SET GLOBAL max_recursive_iterations=2000;
DROP DATABASE IF EXISTS trees_users;
CREATE DATABASE trees_users
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE trees_users;

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

DROP USER IF EXISTS 'trees_users'@'%';
CREATE USER 'trees_users'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON trees_users.* TO 'trees_users'@'%';
GRANT INSERT ON trees_users.* TO 'trees_users'@'%';
GRANT UPDATE ON trees_users.* TO 'trees_users'@'%';
GRANT DELETE ON trees_users.* TO 'trees_users'@'%';
GRANT EXECUTE ON trees_users.* TO 'trees_users'@'%';