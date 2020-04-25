DROP DATABASE IF EXISTS data_pandemic;
CREATE DATABASE data_pandemic;
USE data_pandemic;

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

DROP USER IF EXISTS 'data_pandemic'@'%';
CREATE USER 'data_pandemic'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_pandemic.* TO 'data_pandemic'@'%';
GRANT INSERT ON data_pandemic.* TO 'data_pandemic'@'%';
GRANT UPDATE ON data_pandemic.* TO 'data_pandemic'@'%';
GRANT DELETE ON data_pandemic.* TO 'data_pandemic'@'%';
GRANT EXECUTE ON data_pandemic.* TO 'data_pandemic'@'%';