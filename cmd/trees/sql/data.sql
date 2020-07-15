DROP DATABASE IF EXISTS data_trees;
CREATE DATABASE data_trees;
USE data_trees;

#BIGINT UNSIGNED duration values are all in units of minutes
#BIGINT UNSIGNED fileSize values are all in units of bytes



DROP USER IF EXISTS 'data_trees'@'%';
CREATE USER 'data_trees'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_trees.* TO 'data_trees'@'%';
GRANT INSERT ON data_trees.* TO 'data_trees'@'%';
GRANT UPDATE ON data_trees.* TO 'data_trees'@'%';
GRANT DELETE ON data_trees.* TO 'data_trees'@'%';
GRANT EXECUTE ON data_trees.* TO 'data_trees'@'%';