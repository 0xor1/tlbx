DROP DATABASE IF EXISTS todo_pwds;
CREATE DATABASE todo_pwds
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE todo_pwds;

DROP TABLE IF EXISTS pwds;
CREATE TABLE pwds(
	id BINARY(16) NOT NULL,
	salt   VARBINARY(256) NOT NULL,
	pwd    VARBINARY(256) NOT NULL,
	n      MEDIUMINT UNSIGNED NOT NULL,
	r      MEDIUMINT UNSIGNED NOT NULL,
	p      MEDIUMINT UNSIGNED NOT NULL,
    PRIMARY KEY (id)
);

DROP USER IF EXISTS 'todo_pwds'@'%';
CREATE USER 'todo_pwds'@'%' IDENTIFIED BY 'C0-Mm-0n-Pwd5';
GRANT SELECT ON todo_pwds.* TO 'todo_pwds'@'%';
GRANT INSERT ON todo_pwds.* TO 'todo_pwds'@'%';
GRANT UPDATE ON todo_pwds.* TO 'todo_pwds'@'%';
GRANT DELETE ON todo_pwds.* TO 'todo_pwds'@'%';
GRANT EXECUTE ON todo_pwds.* TO 'todo_pwds'@'%';