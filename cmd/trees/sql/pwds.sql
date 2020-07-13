DROP DATABASE IF EXISTS pwds_trees;
CREATE DATABASE pwds_trees;
USE pwds_trees;

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

DROP USER IF EXISTS 'pwds_trees'@'%';
CREATE USER 'pwds_trees'@'%' IDENTIFIED BY 'C0-Mm-0n-Pwd5';
GRANT SELECT ON pwds_trees.* TO 'pwds_trees'@'%';
GRANT INSERT ON pwds_trees.* TO 'pwds_trees'@'%';
GRANT UPDATE ON pwds_trees.* TO 'pwds_trees'@'%';
GRANT DELETE ON pwds_trees.* TO 'pwds_trees'@'%';
GRANT EXECUTE ON pwds_trees.* TO 'pwds_trees'@'%';