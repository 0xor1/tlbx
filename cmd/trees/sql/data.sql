DROP DATABASE IF EXISTS trees_data;
CREATE DATABASE trees_data
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE trees_data;

#BIGINT UNSIGNED duration values are all in units of minutes
#BIGINT UNSIGNED fileSize values are all in units of bytes

DROP TABLE IF EXISTS users;
CREATE TABLE users(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    handle VARCHAR(15) NOT NULL,
    alias VARCHAR(50) NOT NULL,
    hasAvatar BOOL NOT NULL,
    isActive BOOL NOT NULL DEFAULT 1,
    estimatedTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
    loggedTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
    estimatedExpense BIGINT UNSIGNED NOT NULL DEFAULT 0,
    loggedExpense BIGINT UNSIGNED NOT NULL DEFAULT 0,
    fileCount BIGINT UNSIGNED NOT NULL DEFAULT 0,
    fileSize BIGINT UNSIGNED NOT NULL DEFAULT 0,
    role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
    PRIMARY KEY (host, project, isActive, role, handle),
    UNIQUE INDEX (host, project, isActive, handle, role),
    UNIQUE INDEX (host, project, id),
    UNIQUE INDEX (id, project)
);

DROP TABLE IF EXISTS projectActivities;
CREATE TABLE projectActivities(
	host BINARY(16) NOT NULL,
    project BINARY(16) NOT NULL,
    occurredOn DATETIME(3) NOT NULL,
    user BINARY(16) NOT NULL,
    item BINARY(16) NOT NULL,
    itemType VARCHAR(50) NOT NULL,
    itemHasBeenDeleted BOOL NOT NULL,
    action VARCHAR(50) NOT NULL,
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
    createdOn DATETIME(3) NOT NULL,
    currencyCode VARCHAR(3) NOT NULL,
    hoursPerDay TINYINT UNSIGNED NOT NULL,
    daysPerWeek TINYINT UNSIGNED NOT NULL,
    startOn DATETIME(3) NULL,
    dueOn DATETIME(3) NULL,
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
    createdBy BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    minimumRemainingTime BIGINT UNSIGNED NOT NULL,
    estimatedTime BIGINT UNSIGNED NOT NULL,
    loggedTime BIGINT UNSIGNED NOT NULL,
    estimatedSubTime BIGINT UNSIGNED NOT NULL,
    loggedSubTime BIGINT UNSIGNED NOT NULL,
    estimatedExpense BIGINT UNSIGNED NOT NULL,
    loggedExpense BIGINT UNSIGNED NOT NULL,
    estimatedSubExpense BIGINT UNSIGNED NOT NULL,
    loggedSubExpense BIGINT UNSIGNED NOT NULL,
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

DROP TABLE IF EXISTS times;
CREATE TABLE times(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    task BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdBy BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    taskHasBeenDeleted BOOL NOT NULL,
    taskName VARCHAR(250) NOT NULL,
    duration BIGINT UNSIGNED NOT NULL,
    note VARCHAR(250) NULL,
    PRIMARY KEY(host, project, task, createdOn, createdBy),
    UNIQUE INDEX(host, project, task, id),
    UNIQUE INDEX(host, project, createdBy, createdOn, task),
    UNIQUE INDEX(host, project, createdOn, createdBy, task)
);

DROP TABLE IF EXISTS expenses;
CREATE TABLE expenses(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    task BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdBy BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    taskHasBeenDeleted BOOL NOT NULL,
    taskName VARCHAR(250) NOT NULL,
    paidOn DATETIME(3) NOT NULL,
    value BIGINT UNSIGNED NOT NULL,
    note VARCHAR(250) NULL,
    PRIMARY KEY(host, project, task, createdOn, createdBy),
    UNIQUE INDEX(host, project, task, id),
    UNIQUE INDEX(host, project, createdBy, createdOn, task),
    UNIQUE INDEX(host, project, createdOn, createdBy, task)
);

DROP TABLE IF EXISTS files;
CREATE TABLE files(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    task BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdBy BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    size BIGINT UNSIGNED NOT NULL,
    taskHasBeenDeleted BOOL NOT NULL,
    taskName VARCHAR(250) NOT NULL,
    note VARCHAR(250) NULL,
    PRIMARY KEY(host, project, task, createdOn, id),
    UNIQUE INDEX(host, project, task, id),
    UNIQUE INDEX(host, project, task, createdBy, id)
);

DROP TABLE IF EXISTS comments;
CREATE TABLE comments(
	host BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    task BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    createdBy BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
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