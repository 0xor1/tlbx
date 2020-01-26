DROP DATABASE IF EXISTS data;
CREATE DATABASE data;
USE data;

DROP TABLE IF EXISTS data;
CREATE TABLE data (
    id BINARY(16) NOT NULL,
    PRIMARY KEY name (id)
);

DROP USER IF EXISTS 'data'@'%';
CREATE USER 'data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data.* TO 'data'@'%';
GRANT INSERT ON data.* TO 'data'@'%';
GRANT UPDATE ON data.* TO 'data'@'%';
GRANT DELETE ON data.* TO 'data'@'%';
GRANT EXECUTE ON data.* TO 'data'@'%';