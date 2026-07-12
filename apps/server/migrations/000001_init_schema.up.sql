-- ACM Hot 100 Initial Schema
-- All timestamps use datetime(6) for microsecond precision.
-- All UUIDs use char(36).
-- Engine: InnoDB, Charset: utf8mb4

CREATE TABLE IF NOT EXISTS `users` (
    `id`                CHAR(36)        NOT NULL,
    `email`             VARCHAR(320)    NOT NULL,
    `username`          VARCHAR(32)     NOT NULL,
    `password_hash`     TEXT            NOT NULL,
    `email_verified_at` DATETIME(6)     NULL,
    `status`            VARCHAR(20)     NOT NULL DEFAULT 'PENDING',
    `created_at`        DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    `updated_at`        DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_users_email` (`email`),
    UNIQUE INDEX `idx_users_username` (`username`),
    INDEX `idx_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `problems` (
    `id`                CHAR(36)        NOT NULL,
    `slug`              VARCHAR(80)     NOT NULL,
    `order_index`       INT             NOT NULL,
    `title`             VARCHAR(120)    NOT NULL,
    `difficulty`        VARCHAR(20)     NOT NULL,
    `stage`             VARCHAR(40)     NOT NULL,
    `statement_md`      TEXT            NOT NULL,
    `input_format_md`   TEXT            NOT NULL,
    `output_format_md`  TEXT            NOT NULL,
    `constraints_md`    TEXT            NOT NULL,
    `hints_md`          TEXT            NULL,
    `time_limit_ms`     INT             NOT NULL DEFAULT 1000,
    `memory_limit_kb`   INT             NOT NULL DEFAULT 262144,
    `is_published`      TINYINT(1)      NOT NULL DEFAULT 0,
    `created_at`        DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    `updated_at`        DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_problems_slug` (`slug`),
    UNIQUE INDEX `idx_problems_order` (`order_index`),
    INDEX `idx_problems_difficulty` (`difficulty`),
    INDEX `idx_problems_published` (`is_published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tags` (
    `id`      CHAR(36)      NOT NULL,
    `slug`    VARCHAR(40)   NOT NULL,
    `name`    VARCHAR(60)   NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_tags_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `problem_tags` (
    `problem_id`  CHAR(36)      NOT NULL,
    `tag_id`      CHAR(36)      NOT NULL,
    PRIMARY KEY (`problem_id`, `tag_id`),
    INDEX `idx_problem_tags_tag` (`tag_id`),
    CONSTRAINT `fk_problem_tags_problem` FOREIGN KEY (`problem_id`) REFERENCES `problems` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_problem_tags_tag` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `test_cases` (
    `id`                CHAR(36)        NOT NULL,
    `problem_id`        CHAR(36)        NOT NULL,
    `order_index`       INT             NOT NULL,
    `input_data`        TEXT            NOT NULL,
    `expected_output`   TEXT            NOT NULL,
    `is_sample`         TINYINT(1)      NOT NULL DEFAULT 0,
    `explanation_md`    TEXT            NULL,
    `created_at`        DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_test_cases_problem_order` (`problem_id`, `order_index`),
    CONSTRAINT `fk_test_cases_problem` FOREIGN KEY (`problem_id`) REFERENCES `problems` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `language_configs` (
    `key`                    VARCHAR(20)   NOT NULL,
    `display_name`           VARCHAR(40)   NOT NULL,
    `judge0_language_name`   VARCHAR(120)  NOT NULL,
    `judge0_language_id`     INT           NULL,
    `editor_language`        VARCHAR(30)   NOT NULL,
    `source_template`        TEXT          NOT NULL,
    `enabled`                TINYINT(1)    NOT NULL DEFAULT 1,
    PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `drafts` (
    `user_id`       CHAR(36)       NOT NULL,
    `problem_id`    CHAR(36)       NOT NULL,
    `language_key`  VARCHAR(20)    NOT NULL,
    `source_code`   TEXT           NOT NULL,
    `updated_at`    DATETIME(6)    NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`user_id`, `problem_id`, `language_key`),
    CONSTRAINT `fk_drafts_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_drafts_problem` FOREIGN KEY (`problem_id`) REFERENCES `problems` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_drafts_language` FOREIGN KEY (`language_key`) REFERENCES `language_configs` (`key`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `submissions` (
    `id`                  CHAR(36)        NOT NULL,
    `user_id`             CHAR(36)        NOT NULL,
    `problem_id`          CHAR(36)        NOT NULL,
    `language_key`        VARCHAR(20)     NOT NULL,
    `source_code`         TEXT            NOT NULL,
    `status`              VARCHAR(20)     NOT NULL DEFAULT 'PENDING',
    `passed_cases`        INT             NOT NULL DEFAULT 0,
    `total_cases`         INT             NOT NULL DEFAULT 0,
    `time_ms`             INT             NULL,
    `memory_kb`           INT             NULL,
    `compiler_output`     TEXT            NULL,
    `error_message`       TEXT            NULL,
    `stream_message_id`   VARCHAR(64)     NULL,
    `enqueued_at`         DATETIME(6)     NULL,
    `claimed_at`          DATETIME(6)     NULL,
    `retry_count`         INT             NOT NULL DEFAULT 0,
    `created_at`          DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    `judged_at`           DATETIME(6)     NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_submissions_user` (`user_id`),
    INDEX `idx_submissions_problem` (`problem_id`),
    INDEX `idx_submissions_status` (`status`),
    INDEX `idx_submissions_user_problem` (`user_id`, `problem_id`),
    CONSTRAINT `fk_submissions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_submissions_problem` FOREIGN KEY (`problem_id`) REFERENCES `problems` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_submissions_language` FOREIGN KEY (`language_key`) REFERENCES `language_configs` (`key`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `submission_case_results` (
    `id`                CHAR(36)        NOT NULL,
    `submission_id`     CHAR(36)        NOT NULL,
    `case_index`        INT             NOT NULL,
    `status`            VARCHAR(20)     NOT NULL,
    `time_ms`           INT             NULL,
    `memory_kb`         INT             NULL,
    `actual_output`     TEXT            NULL,
    `expected_output`   TEXT            NULL,
    `is_sample`         TINYINT(1)      NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    INDEX `idx_submission_case_results_submission` (`submission_id`),
    CONSTRAINT `fk_submission_case_results_submission` FOREIGN KEY (`submission_id`) REFERENCES `submissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_problem_progress` (
    `user_id`            CHAR(36)       NOT NULL,
    `problem_id`         CHAR(36)       NOT NULL,
    `state`              VARCHAR(20)    NOT NULL DEFAULT 'NOT_STARTED',
    `attempt_count`      INT            NOT NULL DEFAULT 0,
    `first_ac_at`        DATETIME(6)    NULL,
    `last_submitted_at`  DATETIME(6)    NULL,
    PRIMARY KEY (`user_id`, `problem_id`),
    INDEX `idx_user_problem_progress_state` (`state`),
    CONSTRAINT `fk_user_problem_progress_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_problem_progress_problem` FOREIGN KEY (`problem_id`) REFERENCES `problems` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
