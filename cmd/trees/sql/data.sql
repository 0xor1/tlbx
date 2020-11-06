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
  role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
  PRIMARY KEY (host, project, isActive, role, handle),
  UNIQUE INDEX (host, project, isActive, handle, role),
  UNIQUE INDEX (host, project, id),
  UNIQUE INDEX (id, project),
  UNIQUE INDEX (id, isActive, host, project)
);

DROP TABLE IF EXISTS activities;
CREATE TABLE activities(
  host BINARY(16) NOT NULL,
  project BINARY(16) NOT NULL,
  task BINARY(16) NULL,
  occurredOn DATETIME(3) NOT NULL,
  user BINARY(16) NOT NULL,
  item BINARY(16) NOT NULL,
  itemType VARCHAR(50) NOT NULL,
  itemHasBeenDeleted BOOL NOT NULL,
  action VARCHAR(50) NOT NULL,
  taskName VARCHAR(250) NULL,
  itemName VARCHAR(250) NULL,
  extraInfo VARCHAR(10000) NULL,
  PRIMARY KEY (host, project, occurredOn, item, user),
  UNIQUE INDEX (host, project, item, occurredOn, user),
  UNIQUE INDEX (host, project, task, item, occurredOn, user),
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
  UNIQUE INDEX(id),
  UNIQUE INDEX(host, isArchived, isPublic, name, createdOn, id),
  UNIQUE INDEX(host, isArchived, isPublic, createdOn, name, id),
  UNIQUE INDEX(host, isArchived, isPublic, startOn, name, id),
  UNIQUE INDEX(host, isArchived, isPublic, dueOn, name, id)
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
  description VARCHAR(1250) NOT NULL,
  createdBy BINARY(16) NOT NULL,
  createdOn DATETIME(3) NOT NULL,
  minimumTime BIGINT UNSIGNED NOT NULL,
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
  fileSubCount BIGINT UNSIGNED NOT NULL,
  fileSubSize BIGINT UNSIGNED NOT NULL,
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
  duration BIGINT UNSIGNED NOT NULL,
  note VARCHAR(250) NOT NULL,
  PRIMARY KEY(host, project, task, createdOn, createdBy),
  UNIQUE INDEX(host, project, id),
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
  value BIGINT UNSIGNED NOT NULL,
  note VARCHAR(250) NOT NULL,
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
  isFinalized BOOL NOT NULL,
  id BINARY(16) NOT NULL,
  name VARCHAR(250) NOT NULL,
  createdBy BINARY(16) NOT NULL,
  createdOn DATETIME(3) NOT NULL,
  size BIGINT UNSIGNED NOT NULL,
  mimeType VARCHAR(250) NOT NULL,
  PRIMARY KEY(host, project, task, isFinalized, createdOn, createdBy),
  UNIQUE INDEX(host, project, isFinalized, task, id),
  UNIQUE INDEX(host, project, isFinalized, createdBy, createdOn, task),
  UNIQUE INDEX(host, project, isFinalized, createdOn, createdBy, task),
  UNIQUE INDEX(host, project, isFinalized, name, createdOn, createdBy, task)
);

DROP TABLE IF EXISTS comments;
CREATE TABLE comments(
  host BINARY(16) NOT NULL,
  project BINARY(16) NOT NULL,
  task BINARY(16) NOT NULL,
  id BINARY(16) NOT NULL,
  createdBy BINARY(16) NOT NULL,
  createdOn DATETIME(3) NOT NULL,
  body VARCHAR(10000) NOT NULL,
  PRIMARY KEY(host, project, task, createdOn, createdBy),
  UNIQUE INDEX(host, project, task, id),
  UNIQUE INDEX(host, project, createdBy, createdOn, task),
  UNIQUE INDEX(host, project, createdOn, createdBy, task)
);

#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!#
#********************************MAGIC PROCEDURE WARNING*********************************#
# THIS PROCEDURE MUST ONLY BE CALLED INTERNALLY BY 
# taskeps.go func setAncestralChainAggregateValuesFromTask                                                                                           #
#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!#
DROP PROCEDURE IF EXISTS setAncestralChainAggregateValuesFromTask;
DELIMITER //
CREATE PROCEDURE setAncestralChainAggregateValuesFromTask(_host BINARY(16), _project BINARY(16), _task BINARY(16))
BEGIN
  
  DECLARE currentParent BINARY(16) DEFAULT NULL;
  DECLARE currentMinimumTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentEstimatedSubTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentLoggedSubTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentEstimatedSubExpense BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentLoggedSubExpense BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentFileSubCount BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentFileSubSize BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentChildCount BIGINT UNSIGNED DEFAULT 0;
  DECLARE currentDescendantCount BIGINT UNSIGNED DEFAULT 0;

  DECLARE newMinimumTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE newEstimatedSubTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE newLoggedSubTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE newEstimatedSubExpense BIGINT UNSIGNED DEFAULT 0;
  DECLARE newLoggedSubExpense BIGINT UNSIGNED DEFAULT 0;
  DECLARE newFileSubCount BIGINT UNSIGNED DEFAULT 0;
  DECLARE newFileSubSize BIGINT UNSIGNED DEFAULT 0;
  DECLARE newChildCount BIGINT UNSIGNED DEFAULT 0;
  DECLARE newDescendantCount BIGINT UNSIGNED DEFAULT 0;

  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  CREATE TEMPORARY TABLE tempUpdatedIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );

  WHILE _task IS NOT NULL DO
    
    SELECT
      t.parent,
      t.minimumTime,
      t.estimatedSubTime,
      t.loggedSubTime,
      t.estimatedSubExpense,
      t.loggedSubExpense,
      t.fileSubCount,
      t.fileSubSize,
      t.childCount,
      t.descendantCount,
      t.estimatedTime + CASE t.isParallel
        WHEN 0 THEN COALESCE(Sum(c.minimumTime), 0)
        WHEN 1 THEN COALESCE(MAX(c.minimumTime), 0)
      END,
      COALESCE(SUM(c.estimatedTime + c.estimatedSubTime), 0),
      COALESCE(SUM(c.loggedTime + c.loggedSubTime), 0),
      COALESCE(SUM(c.estimatedExpense + c.estimatedSubExpense), 0),
      COALESCE(SUM(c.loggedExpense + c.loggedSubExpense), 0),
      COALESCE(SUM(c.fileCount + c.fileSubCount), 0),
      COALESCE(SUM(c.fileSize + c.fileSubSize), 0),
      COALESCE(COUNT(DISTINCT c.id), 0),
      COALESCE(COALESCE(COUNT(DISTINCT c.id), 0) + COALESCE(SUM(c.descendantCount), 0), 0)
    INTO
      currentParent,
      currentMinimumTime,
      currentEstimatedSubTime,
      currentLoggedSubTime,
      currentEstimatedSubExpense,
      currentLoggedSubExpense,
      currentFileSubCount,
      currentFileSubSize,
      currentChildCount,
      currentDescendantCount,
      newMinimumTime,
      newEstimatedSubTime,
      newLoggedSubTime,
      newEstimatedSubExpense,
      newLoggedSubExpense,
      newFileSubCount,
      newFileSubSize,
      newChildCount,
      newDescendantCount
    FROM
      tasks t
    LEFT JOIN
      tasks c
    ON
      c.host=_host
    AND
      c.project=_project
    AND
      c.parent=_task
    WHERE
      t.host=_host
    AND
      t.project=_project
    AND
      t.id=_task
    GROUP BY
      t.id;

    IF currentMinimumTime <> newMinimumTime OR
      currentEstimatedSubTime <> newEstimatedSubTime OR
      currentLoggedSubTime <> newLoggedSubTime OR
      currentEstimatedSubExpense <> newEstimatedSubExpense OR
      currentLoggedSubExpense <> newLoggedSubExpense OR
      currentFileSubCount <> newFileSubCount OR
      currentFileSubSize <> newFileSubSize OR
      currentChildCount <> newChildCount OR
      currentDescendantCount <> newDescendantCount THEN

      UPDATE
        tasks
      SET
        minimumTime=newMinimumTime,
        estimatedSubTime=newEstimatedSubTime,
        loggedSubTime=newLoggedSubTime,
        estimatedSubExpense=newEstimatedSubExpense,
        loggedSubExpense=newLoggedSubExpense,
        fileSubCount=newFileSubCount,
        fileSubSize=newFileSubSize,
        childCount=newChildCount,
        descendantCount=newDescendantCount
      WHERE
        host=_host
      AND
        project=_project
      AND
        id=_task;

      INSERT INTO tempUpdatedIds VALUES (_task);
      
      SET _task = currentParent;
    
    ELSE

      SET _task = NULL;

    END IF;

  END WHILE;

  SELECT id FROM tempUpdatedIds;
END //
DELIMITER ;

#useful helper query for manual verifying/debugging test results
#SELECT  t1.name, t2.name AS parent, t3.name AS nextSibling, t4.name AS firstChild, t1.description, t1.minimumTime, t1.estimatedTime, t1.loggedTime, t1.estimatedSubTime, t1.loggedSubTime, t1.estimatedExpense, t1.loggedExpense, t1.estimatedSubExpense, t1.loggedSubExpense, t1.childCount, t1.descendantCount FROM trees_data.tasks t1 LEFT JOIN trees_data.tasks t2 ON t1.parent = t2.id LEFT JOIN trees_data.tasks t3 ON t1.nextSibling = t3.id LEFT JOIN trees_data.tasks t4 ON t1.firstChild = t4.id ORDER BY t1.name;


DROP USER IF EXISTS 'trees_data'@'%';
CREATE USER 'trees_data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON trees_data.* TO 'trees_data'@'%';
GRANT INSERT ON trees_data.* TO 'trees_data'@'%';
GRANT UPDATE ON trees_data.* TO 'trees_data'@'%';
GRANT DELETE ON trees_data.* TO 'trees_data'@'%';
GRANT EXECUTE ON trees_data.* TO 'trees_data'@'%';