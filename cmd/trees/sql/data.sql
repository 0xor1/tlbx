DROP DATABASE IF EXISTS trees_data;
CREATE DATABASE trees_data
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE trees_data;

#BIGINT UNSIGNED duration values are all in units of minutes
#BIGINT UNSIGNED fileSize values are all in units of bytes

DROP TABLE IF EXISTS projectUsers;
CREATE TABLE projectUsers(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    alias VARCHAR(50) NOT NULL,
    isActive BOOL NOT NULL,
    totalEstimatedTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
    PRIMARY KEY (host, project, isActive, role, alias),
    UNIQUE INDEX (host, project, isActive, alias, role),
    UNIQUE INDEX (host, project, id),
    UNIQUE INDEX (host, id, project)
);

DROP TABLE IF EXISTS projectActivities;
CREATE TABLE projectActivities(
	host BINARY(16) NOT NULL,
    project BINARY(16) NOT NULL,
    occurredOn DATETIME NOT NULL,
    user BINARY(16) NOT NULL,
    item BINARY(16) NOT NULL,
    itemType VARCHAR(100) NOT NULL,
    itemHasBeenDeleted BOOL NOT NULL,
    action VARCHAR(100) NOT NULL,
    itemName VARCHAR(250) NULL,
    extraInfo VARCHAR(1250) NULL,
    PRIMARY KEY (host, project, occurredOn, item, user),
    UNIQUE INDEX (host, project, item, occurredOn, user),
    UNIQUE INDEX (host, project, user, occurredOn, item)
);

DROP TABLE IF EXISTS projectLocks;
CREATE TABLE projectLocks(
	host BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    PRIMARY KEY(host, id)
);

DROP TABLE IF EXISTS projects;
CREATE TABLE projects(
	host BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    isArchived BOOL NOT NULL,
	name VARCHAR(250) NOT NULL,
    createdOn DATETIME NOT NULL,
    currencyCode VARCHAR NOT NULL, 
    hoursPerDay TINYINT UNSIGNED NOT NULL,
    daysPerWeek TINYINT UNSIGNED NOT NULL,
    startOn DATETIME NULL,
    dueOn DATETIME NULL,
    isPublic BOOL NOT NULL,
    PRIMARY KEY (host, id),
    INDEX(host, isArchived, isPublic, name, createdOn, id),
    INDEX(host, isArchived, isPublic, createdOn, name, id),
    INDEX(host, isArchived, isPublic, startOn, name, id),
    INDEX(host, isArchived, isPublic, dueOn, name, id)
);

DROP TABLE IF EXISTS tasks;
CREATE TABLE tasks(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    parent BINARY(16) NULL,
    firstChild BINARY(16) NULL,
    nextSibling BINARY(16) NULL,
    user BINARY(16) NULL,
	name VARCHAR(250) NOT NULL,
	description VARCHAR(1250) NULL,
    createdOn DATETIME NOT NULL,
    minimumRemainingTime BIGINT UNSIGNED NOT NULL,
    estimatedTime BIGINT UNSIGNED NOT NULL,
    loggedTime BIGINT UNSIGNED NOT NULL,
    estimatedSubTime BIGINT UNSIGNED NOT NULL,
    loggedSubTime BIGINT UNSIGNED NOT NULL,
    estimatedCost DECIMAL(13,4) UNSIGNED NOT NULL,
    loggedCost DECIMAL(13,4) UNSIGNED NOT NULL,
    estimatedSubCost DECIMAL(13,4) UNSIGNED NOT NULL,
    loggedSubCost DECIMAL(13,4) UNSIGNED NOT NULL,
    fileCount BIGINT UNSIGNED NOT NULL,
    fileSize BIGINT UNSIGNED NOT NULL,
    subFileCount BIGINT UNSIGNED NOT NULL,
    subFileSize BIGINT UNSIGNED NOT NULL,
    childCount BIGINT UNSIGNED NOT NULL,
    descendantCount BIGINT UNSIGNED NOT NULL,
    isParallel BOOL NOT NULL,
    PRIMARY KEY (host, project, id),
    UNIQUE INDEX(host, user, id),
    UNIQUE INDEX(host, project, parent, id),
    UNIQUE INDEX(host, project, nextSibling, id),
    UNIQUE INDEX(host, project, user, id)
);

DROP TABLE IF EXISTS timeLogs;
CREATE TABLE timeLogs(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    task BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    loggedBy BINARY(16) NOT NULL,
    loggedOn DATETIME NOT NULL,
    taskHasBeenDeleted BOOL NOT NULL,
    taskName VARCHAR(250) NOT NULL,
    duration BIGINT UNSIGNED NOT NULL,
    note VARCHAR(250) NULL,
    PRIMARY KEY(host, project, task, loggedOn, loggedBy),
    UNIQUE INDEX(host, project, task, id),
    UNIQUE INDEX(host, project, loggedBy, loggedOn, task),
    UNIQUE INDEX(host, project, loggedOn, loggedBy, task)
);

DROP TABLE IF EXISTS files;
CREATE TABLE files(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    task BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    uploadedBy BINARY(16) NOT NULL,
    uploadedOn DATETIME NOT NULL,
    size BIGINT UNSIGNED NOT NULL,
    taskHasBeenDeleted BOOL NOT NULL,
    taskName VARCHAR(250) NOT NULL,
    note VARCHAR(250) NULL,
    PRIMARY KEY(host, project, task, uploadedOn, id),
    UNIQUE INDEX(host, project, task, id),
    UNIQUE INDEX(host, project, task, uploadedBy, id)
);

DROP TABLE IF EXISTS comments;
CREATE TABLE comments(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    task BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdBy BINARY(16) NOT NULL,
    createdOn DATETIME NOT NULL,
    body VARCHAR(10000),
    taskHasBeenDeleted BOOL NOT NULL,
    taskName VARCHAR(250) NOT NULL,
    note VARCHAR(250) NULL,
    PRIMARY KEY(host, project, task, createdOn, id),
    UNIQUE INDEX(host, project, task, id),
    UNIQUE INDEX(host, project, task, createdBy, id)
);

DROP USER IF EXISTS 'trees_data'@'%';
CREATE USER 'trees_data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON trees_data.* TO 'trees_data'@'%';
GRANT INSERT ON trees_data.* TO 'trees_data'@'%';
GRANT UPDATE ON trees_data.* TO 'trees_data'@'%';
GRANT DELETE ON trees_data.* TO 'trees_data'@'%';
GRANT EXECUTE ON trees_data.* TO 'trees_data'@'%';