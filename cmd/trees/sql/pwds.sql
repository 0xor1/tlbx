DROP DATABASE IF EXISTS trees_pwds;
CREATE DATABASE trees_pwds
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE trees_pwds;

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

DROP USER IF EXISTS 'trees_pwds'@'%';
CREATE USER 'trees_pwds'@'%' IDENTIFIED BY 'C0-Mm-0n-Pwd5';
GRANT SELECT ON trees_pwds.* TO 'trees_pwds'@'%';
GRANT INSERT ON trees_pwds.* TO 'trees_pwds'@'%';
GRANT UPDATE ON trees_pwds.* TO 'trees_pwds'@'%';
GRANT DELETE ON trees_pwds.* TO 'trees_pwds'@'%';
GRANT EXECUTE ON trees_pwds.* TO 'trees_pwds'@'%';