DROP DATABASE IF EXISTS data_games;
CREATE DATABASE data_games;
USE data_games;

DROP TABLE IF EXISTS games;
CREATE TABLE games (
    id BINARY(16) NOT NULL,
    type VARCHAR(20) NOT NULL,
    updatedOn DATETIME(3) NOT NULL,
    serialized VARBINARY(5000) NOT NULL,
    PRIMARY KEY id (id, type),
    UNIQUE INDEX updatedOn (updatedOn, id, type)
);

DROP TABLE IF EXISTS players;
CREATE TABLE players (
    id BINARY(16) NOT NULL,
    game BINARY(16) NOT NULL,
    PRIMARY KEY id (id),
    UNIQUE INDEX game (game, id),
    FOREIGN KEY (game) REFERENCES games (id) ON DELETE CASCADE
);

DROP USER IF EXISTS 'data_games'@'%';
CREATE USER 'data_games'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data_games.* TO 'data_games'@'%';
GRANT INSERT ON data_games.* TO 'data_games'@'%';
GRANT UPDATE ON data_games.* TO 'data_games'@'%';
GRANT DELETE ON data_games.* TO 'data_games'@'%';
GRANT EXECUTE ON data_games.* TO 'data_games'@'%';