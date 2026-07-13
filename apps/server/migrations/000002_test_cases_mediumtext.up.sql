-- Change test_cases input_data and expected_output from TEXT to MEDIUMTEXT
-- to support large test case data (TEXT is limited to ~64KB, MEDIUMTEXT supports ~16MB)

ALTER TABLE `test_cases` MODIFY COLUMN `input_data` MEDIUMTEXT NOT NULL;
ALTER TABLE `test_cases` MODIFY COLUMN `expected_output` MEDIUMTEXT NOT NULL;
