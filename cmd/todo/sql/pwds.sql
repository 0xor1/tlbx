DROP DATABASE IF EXISTS pwds_todo;
CREATE DATABASE pwds_todo;
USE pwds_todo;

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

DROP USER IF EXISTS 'pwds_todo'@'%';
CREATE USER 'pwds_todo'@'%' IDENTIFIED BY 'C0-Mm-0n-Pwd5';
GRANT SELECT ON pwds_todo.* TO 'pwds_todo'@'%';
GRANT INSERT ON pwds_todo.* TO 'pwds_todo'@'%';
GRANT UPDATE ON pwds_todo.* TO 'pwds_todo'@'%';
GRANT DELETE ON pwds_todo.* TO 'pwds_todo'@'%';
GRANT EXECUTE ON pwds_todo.* TO 'pwds_todo'@'%';