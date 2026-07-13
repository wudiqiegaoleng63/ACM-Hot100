-- Revert test_cases input_data and expected_output back to TEXT

ALTER TABLE `test_cases` MODIFY COLUMN `input_data` TEXT NOT NULL;
ALTER TABLE `test_cases` MODIFY COLUMN `expected_output` TEXT NOT NULL;
