DROP DATABASE IF EXISTS trees_data;
CREATE DATABASE trees_data
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE trees_data;

#BIGINT UNSIGNED time values are all in units of minutes
#BIGINT UNSIGNED fileSize values are all in units of bytes
## abbreviations used est=estimate inc=incurred sub=subtask(s) prev=previous sib=siblings and N suffix means count

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
  UNIQUE INDEX (host, project, itemHasBeenDeleted, occurredOn, item, user),
  UNIQUE INDEX (host, project, item, occurredOn, user),
  UNIQUE INDEX (host, project, task, item, occurredOn, user),
  UNIQUE INDEX (host, project, user, occurredOn, item),
  UNIQUE INDEX (host, project, user, itemHasBeenDeleted, occurredOn, item)
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
  hoursPerDay TINYINT UNSIGNED NULL,
  daysPerWeek TINYINT UNSIGNED NULL,
  startOn DATETIME(3) NULL,
  endOn DATETIME(3) NULL,
  isPublic BOOL NOT NULL,
  PRIMARY KEY (host, id),
  UNIQUE INDEX(id),
  UNIQUE INDEX(host, isArchived, isPublic, name, createdOn, id),
  UNIQUE INDEX(host, isArchived, isPublic, createdOn, name, id),
  UNIQUE INDEX(host, isArchived, isPublic, startOn, name, id),
  UNIQUE INDEX(host, isArchived, isPublic, endOn, name, id)
);

DROP TABLE IF EXISTS tasks;
CREATE TABLE tasks(
  host BINARY(16) NOT NULL,
  project BINARY(16) NOT NULL,
  id BINARY(16) NOT NULL,
  parent BINARY(16) NULL,
  firstChild BINARY(16) NULL,
  nextSib BINARY(16) NULL,
  user BINARY(16) NULL,
  name VARCHAR(250) NOT NULL,
  description VARCHAR(1250) NOT NULL,
  createdBy BINARY(16) NOT NULL,
  createdOn DATETIME(3) NOT NULL,
  timeEst BIGINT UNSIGNED NOT NULL,
  timeInc BIGINT UNSIGNED NOT NULL,
  timeSubMin BIGINT UNSIGNED NOT NULL,
  timeSubEst BIGINT UNSIGNED NOT NULL,
  timeSubInc BIGINT UNSIGNED NOT NULL,
  costEst BIGINT UNSIGNED NOT NULL,
  costInc BIGINT UNSIGNED NOT NULL,
  costSubEst BIGINT UNSIGNED NOT NULL,
  costSubInc BIGINT UNSIGNED NOT NULL,
  fileN BIGINT UNSIGNED NOT NULL,
  fileSize BIGINT UNSIGNED NOT NULL,
  fileSubN BIGINT UNSIGNED NOT NULL,
  fileSubSize BIGINT UNSIGNED NOT NULL,
  childN BIGINT UNSIGNED NOT NULL,
  descN BIGINT UNSIGNED NOT NULL,
  isParallel BOOL NOT NULL,
  PRIMARY KEY (host, project, id),
  UNIQUE INDEX(host, user, id),
  UNIQUE INDEX(host, project, parent, id),
  UNIQUE INDEX(host, project, nextSib, id),
  UNIQUE INDEX(host, project, user, id)
);

DROP TABLE IF EXISTS vitems;
CREATE TABLE vitems(
  host BINARY(16) NOT NULL,
  project BINARY(16) NOT NULL,
  task BINARY(16) NOT NULL,
  type VARCHAR(50) NOT NULL,
  id BINARY(16) NOT NULL,
  createdBy BINARY(16) NOT NULL,
  createdOn DATETIME(3) NOT NULL,
  inc BIGINT UNSIGNED NOT NULL,
  note VARCHAR(250) NOT NULL,
  PRIMARY KEY(host, project, task, type, createdOn, createdBy),
  UNIQUE INDEX(host, project, type, id),
  UNIQUE INDEX(host, project, task, type, id),
  UNIQUE INDEX(host, project, createdBy, type, createdOn, task),
  UNIQUE INDEX(host, project, type, createdOn, createdBy, task)
);

DROP TABLE IF EXISTS files;
CREATE TABLE files(
  host BINARY(16) NOT NULL,
  project BINARY(16) NOT NULL,
  task BINARY(16) NOT NULL,
  id BINARY(16) NOT NULL,
  name VARCHAR(250) NOT NULL,
  createdBy BINARY(16) NOT NULL,
  createdOn DATETIME(3) NOT NULL,
  size BIGINT UNSIGNED NOT NULL,
  type VARCHAR(250) NOT NULL,
  PRIMARY KEY(host, project, task, createdOn, createdBy),
  UNIQUE INDEX(host, project, task, id),
  UNIQUE INDEX(host, project, createdBy, createdOn, task),
  UNIQUE INDEX(host, project, createdOn, createdBy, task),
  UNIQUE INDEX(host, project, name, createdOn, createdBy, task)
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
  
  DECLARE curParent BINARY(16) DEFAULT NULL;

  DECLARE curTimeSubMin BIGINT UNSIGNED DEFAULT 0;
  DECLARE curTimeSubEst BIGINT UNSIGNED DEFAULT 0;
  DECLARE curTimeSubInc BIGINT UNSIGNED DEFAULT 0;
  DECLARE curCostSubEst BIGINT UNSIGNED DEFAULT 0;
  DECLARE curCostSubInc BIGINT UNSIGNED DEFAULT 0;
  DECLARE curFileSubN BIGINT UNSIGNED DEFAULT 0;
  DECLARE curFileSubSize BIGINT UNSIGNED DEFAULT 0;
  DECLARE curChildN BIGINT UNSIGNED DEFAULT 0;
  DECLARE curDescN BIGINT UNSIGNED DEFAULT 0;

  DECLARE newTimeSubMin BIGINT UNSIGNED DEFAULT 0;
  DECLARE newTimeSubEst BIGINT UNSIGNED DEFAULT 0;
  DECLARE newTimeSubInc BIGINT UNSIGNED DEFAULT 0;
  DECLARE newCostSubEst BIGINT UNSIGNED DEFAULT 0;
  DECLARE newCostSubInc BIGINT UNSIGNED DEFAULT 0;
  DECLARE newFileSubN BIGINT UNSIGNED DEFAULT 0;
  DECLARE newFileSubSize BIGINT UNSIGNED DEFAULT 0;
  DECLARE newChildN BIGINT UNSIGNED DEFAULT 0;
  DECLARE newDescN BIGINT UNSIGNED DEFAULT 0;

  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  CREATE TEMPORARY TABLE tempUpdatedIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );

  WHILE _task IS NOT NULL DO
    
    SELECT
      t.parent,
      t.timeSubMin,
      t.timeSubEst,
      t.timeSubInc,
      t.costSubEst,
      t.costSubInc,
      t.fileSubN,
      t.fileSubSize,
      t.childN,
      t.descN,
      CASE t.isParallel
        WHEN 0 THEN COALESCE(SUM(c.timeEst + c.timeSubMin), 0)
        WHEN 1 THEN COALESCE(MAX(c.timeEst + c.timeSubMin), 0)
      END,
      COALESCE(SUM(c.timeEst + c.timeSubEst), 0),
      COALESCE(SUM(c.timeInc + c.timeSubInc), 0),
      COALESCE(SUM(c.costEst + c.costSubEst), 0),
      COALESCE(SUM(c.costInc + c.costSubInc), 0),
      COALESCE(SUM(c.fileN + c.fileSubN), 0),
      COALESCE(SUM(c.fileSize + c.fileSubSize), 0),
      COALESCE(COUNT(DISTINCT c.id), 0),
      COALESCE(COALESCE(COUNT(DISTINCT c.id), 0) + COALESCE(SUM(c.descN), 0), 0)
    INTO
      curParent,
      curTimeSubMin,
      curTimeSubEst,
      curTimeSubInc,
      curCostSubEst,
      curCostSubInc,
      curFileSubN,
      curFileSubSize,
      curChildN,
      curDescN,
      newtimeSubMin,
      newtimeSubEst,
      newTimeSubInc,
      newCostSubEst,
      newCostSubInc,
      newFileSubN,
      newFileSubSize,
      newChildN,
      newDescN
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

    IF curTimeSubMin <> newtimeSubMin OR
      curTimeSubEst <> newtimeSubEst OR
      curTimeSubInc <> newTimeSubInc OR
      curCostSubEst <> newCostSubEst OR
      curCostSubInc <> newCostSubInc OR
      curFileSubN <> newFileSubN OR
      curFileSubSize <> newFileSubSize OR
      curChildN <> newChildN OR
      curDescN <> newDescN THEN

      UPDATE
        tasks
      SET
        timeSubMin=newtimeSubMin,
        timeSubEst=newtimeSubEst,
        timeSubInc=newTimeSubInc,
        costSubEst=newCostSubEst,
        costSubInc=newCostSubInc,
        fileSubN=newFileSubN,
        fileSubSize=newFileSubSize,
        childN=newChildN,
        descN=newDescN
      WHERE
        host=_host
      AND
        project=_project
      AND
        id=_task;

      INSERT INTO tempUpdatedIds VALUES (_task);
      
      SET _task = curParent;
    
    ELSE

      SET _task = NULL;

    END IF;

  END WHILE;

  SELECT id FROM tempUpdatedIds;
END //
DELIMITER ;

#useful helper query for manual verifying/debugging test results
#SELECT  t1.name, t2.name AS parent, t3.name AS nextSib, t4.name AS firstChild, t1.description, t1.timeEst, t1.timeInc, t1.timeSubMin, t1.timeSubEst, t1.timeSubInc, t1.costEst, t1.costInc, t1.costSubEst, t1.costSubInc, t1.childN, t1.descN FROM trees_data.tasks t1 LEFT JOIN trees_data.tasks t2 ON t1.parent = t2.id LEFT JOIN trees_data.tasks t3 ON t1.nextSib = t3.id LEFT JOIN trees_data.tasks t4 ON t1.firstChild = t4.id ORDER BY t1.name;


DROP USER IF EXISTS 'trees_data'@'%';
CREATE USER 'trees_data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON trees_data.* TO 'trees_data'@'%';
GRANT INSERT ON trees_data.* TO 'trees_data'@'%';
GRANT UPDATE ON trees_data.* TO 'trees_data'@'%';
GRANT DELETE ON trees_data.* TO 'trees_data'@'%';
GRANT EXECUTE ON trees_data.* TO 'trees_data'@'%';