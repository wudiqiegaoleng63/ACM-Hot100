-- Align submission status values with the MVP design spec.
-- Non-terminal: QUEUED, COMPILING, RUNNING
-- Terminal: AC, WA, TLE, MLE, RE, CE, SYSTEM_ERROR

ALTER TABLE `submissions`
  MODIFY COLUMN `status` VARCHAR(20) NOT NULL DEFAULT 'QUEUED';
