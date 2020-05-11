DROP DATABASE IF EXISTS data_boring;
CREATE DATABASE data_boring;
USE data_boring;

DROP TABLE IF EXISTS games;
CREATE TABLE games (
    id BINARY(16) NOT NULL,
    updatedOn DATETIME(3) NOT NULL,
    serialized VARBINARY(5000) NOT NULL,
    PRIMARY KEY id (id),
    UNIQUE INDEX updatedOn (updatedOn, id)
);

DROP TABLE IF EXISTS players;
CREATE TABLE players (
    id BINARY(16) NOT NULL,
    game BINARY(16) NOT NULL,
    PRIMARY KEY id (id),
    UNIQUE INDEX game (game, id)
    FOREIGN KEY (game) REFERENCES games (id) ON DELETE CASCADE
);

DROP USER IF EXISTS 'data_boring'@'%';
CREATE USER 'data_boring'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_boring.* TO 'data_boring'@'%';
GRANT INSERT ON data_boring.* TO 'data_boring'@'%';
GRANT UPDATE ON data_boring.* TO 'data_boring'@'%';
GRANT DELETE ON data_boring.* TO 'data_boring'@'%';
GRANT EXECUTE ON data_boring.* TO 'data_boring'@'%';