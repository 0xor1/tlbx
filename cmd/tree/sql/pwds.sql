DROP DATABASE IF EXISTS pwds_tree;
CREATE DATABASE pwds_tree;
USE pwds_tree;

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

DROP USER IF EXISTS 'pwds_tree'@'%';
CREATE USER 'pwds_tree'@'%' IDENTIFIED BY 'C0-Mm-0n-Pwd5';
GRANT SELECT ON pwds_tree.* TO 'pwds_tree'@'%';
GRANT INSERT ON pwds_tree.* TO 'pwds_tree'@'%';
GRANT UPDATE ON pwds_tree.* TO 'pwds_tree'@'%';
GRANT DELETE ON pwds_tree.* TO 'pwds_tree'@'%';
GRANT EXECUTE ON pwds_tree.* TO 'pwds_tree'@'%';