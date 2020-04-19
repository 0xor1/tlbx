DROP DATABASE IF EXISTS data;
CREATE DATABASE data;
USE data;

DROP TABLE IF EXISTS games;
CREATE TABLE games (
    id BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    pwd VARCHAR(8) NOT NULL,
    state VARBINARY(5000) NOT NULL,
    PRIMARY KEY id (id)
);

DROP TABLE IF EXISTS players;
CREATE TABLE players (
    id BINARY(16) NOT NULL,
    game BINARY(16) NOT NULL,
    PRIMARY KEY id (id)
);

DROP USER IF EXISTS 'data'@'%';
CREATE USER 'data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data.* TO 'data'@'%';
GRANT INSERT ON data.* TO 'data'@'%';
GRANT UPDATE ON data.* TO 'data'@'%';
GRANT DELETE ON data.* TO 'data'@'%';
GRANT EXECUTE ON data.* TO 'data'@'%';