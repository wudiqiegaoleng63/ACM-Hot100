package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/acmhot100/server/internal/model"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─── YAML structures ────────────────────────────────────────────────────────

// problemYAML represents the structure of a problem.yaml file.
type problemYAML struct {
	Slug          string `yaml:"slug"`
	OrderIndex    int    `yaml:"order_index"`
	Title         string `yaml:"title"`
	Difficulty    string `yaml:"difficulty"`
	Stage         string `yaml:"stage"`
	TimeLimitMs   int    `yaml:"time_limit_ms"`
	MemoryLimitKb int    `yaml:"memory_limit_kb"`
	Tags          []struct {
		Slug string `yaml:"slug"`
		Name string `yaml:"name"`
	} `yaml:"tags"`
}

// ─── SeedAll ────────────────────────────────────────────────────────────────

// SeedAll seeds the database with language configs and problems.
func SeedAll(db *gorm.DB, seedDir string) error {
	log.Println("Seeding language configs...")
	if err := SeedLanguageConfigs(db); err != nil {
		return fmt.Errorf("seed language configs: %w", err)
	}

	log.Println("Seeding problems...")
	if err := SeedProblems(db, seedDir); err != nil {
		return fmt.Errorf("seed problems: %w", err)
	}

	log.Println("Seed completed successfully")
	return nil
}

// ─── SeedLanguageConfigs ────────────────────────────────────────────────────

// SeedLanguageConfigs inserts language configurations into the database.
func SeedLanguageConfigs(db *gorm.DB) error {
	configs := []model.LanguageConfig{
		{
			Key:                "cpp17",
			DisplayName:        "C++17",
			Judge0LanguageName: "C++ (GCC 14.1.0)",
			Judge0LanguageID:   intPointer(105),
			EditorLanguage:     "cpp",
			SourceTemplate: `#include <bits/stdc++.h>
using namespace std;

int main() {
    // Write your solution here

    return 0;
}`,
			Enabled: true,
		},
		{
			Key:                "python3",
			DisplayName:        "Python 3",
			Judge0LanguageName: "Python (3.12.5)",
			Judge0LanguageID:   intPointer(100),
			EditorLanguage:     "python",
			SourceTemplate: `# Write your solution here
def main():
    pass

if __name__ == "__main__":
    main()`,
			Enabled: true,
		},
		{
			Key:                "java17",
			DisplayName:        "Java 17",
			Judge0LanguageName: "Java (JDK 17.0.6)",
			Judge0LanguageID:   intPointer(91),
			EditorLanguage:     "java",
			SourceTemplate: `import java.util.*;

public class Main {
    public static void main(String[] args) {
        // Write your solution here
    }
}`,
			Enabled: true,
		},
	}

	for _, cfg := range configs {
		result := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"display_name", "judge0_language_name", "judge0_language_id", "editor_language", "source_template", "enabled"}),
		}).Create(&cfg)

		if result.Error != nil {
			return fmt.Errorf("upsert language config %s: %w", cfg.Key, result.Error)
		}
		log.Printf("  Upserted language config: %s (%s)", cfg.Key, cfg.DisplayName)
	}

	return nil
}

func intPointer(value int) *int {
	return &value
}

// ─── SeedProblems ───────────────────────────────────────────────────────────

// SeedProblems reads problem directories from seedDir and inserts them into the database.
func SeedProblems(db *gorm.DB, seedDir string) error {
	// Read problem directories
	entries, err := os.ReadDir(seedDir)
	if err != nil {
		return fmt.Errorf("read seed directory %s: %w", seedDir, err)
	}

	// Sort directories by name to ensure consistent ordering
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		problemDir := filepath.Join(seedDir, entry.Name())
		if err := seedProblem(db, problemDir); err != nil {
			return fmt.Errorf("seed problem from %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// seedProblem seeds a single problem from its directory.
func seedProblem(db *gorm.DB, problemDir string) error {
	// Parse problem.yaml
	yamlPath := filepath.Join(problemDir, "problem.yaml")
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("read problem.yaml: %w", err)
	}

	var py problemYAML
	if err := yaml.Unmarshal(yamlData, &py); err != nil {
		return fmt.Errorf("parse problem.yaml: %w", err)
	}

	// Read statement.md
	statementPath := filepath.Join(problemDir, "statement.md")
	statementMD, err := os.ReadFile(statementPath)
	if err != nil {
		return fmt.Errorf("read statement.md: %w", err)
	}

	// Parse sections from statement.md
	inputFormat, outputFormat, constraints, hints := parseStatementSections(string(statementMD))

	// Create problem record
	problem := model.Problem{
		ID:             uuid.New().String(),
		Slug:           py.Slug,
		OrderIndex:     py.OrderIndex,
		Title:          py.Title,
		Difficulty:     py.Difficulty,
		Stage:          py.Stage,
		StatementMD:    string(statementMD),
		InputFormatMD:  inputFormat,
		OutputFormatMD: outputFormat,
		ConstraintsMD:  constraints,
		HintsMD:        hints,
		TimeLimitMs:    py.TimeLimitMs,
		MemoryLimitKb:  py.MemoryLimitKb,
		IsPublished:    true,
	}

	// Upsert problem by slug
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"order_index", "title", "difficulty", "stage", "statement_md", "input_format_md", "output_format_md", "constraints_md", "hints_md", "time_limit_ms", "memory_limit_kb", "is_published"}),
	}).Create(&problem)

	if result.Error != nil {
		return fmt.Errorf("upsert problem %s: %w", py.Slug, result.Error)
	}

	// Reload problem to get the actual ID (in case of conflict, we need the existing ID)
	var existingProblem model.Problem
	if err := db.Where("slug = ?", py.Slug).First(&existingProblem).Error; err != nil {
		return fmt.Errorf("find problem %s: %w", py.Slug, err)
	}
	problem.ID = existingProblem.ID

	log.Printf("  Upserted problem: %s (%s)", py.Slug, py.Title)

	// Upsert tags and create associations
	tagIDs := make([]string, 0, len(py.Tags))
	for _, t := range py.Tags {
		tag := model.Tag{
			ID:   uuid.New().String(),
			Slug: t.Slug,
			Name: t.Name,
		}

		// Upsert tag by slug
		result = db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{"name"}),
		}).Create(&tag)

		if result.Error != nil {
			return fmt.Errorf("upsert tag %s: %w", t.Slug, result.Error)
		}

		// Get the actual tag ID
		var existingTag model.Tag
		if err := db.Where("slug = ?", t.Slug).First(&existingTag).Error; err != nil {
			return fmt.Errorf("find tag %s: %w", t.Slug, err)
		}
		tagIDs = append(tagIDs, existingTag.ID)
	}

	// Delete existing problem-tag associations and recreate
	if err := db.Where("problem_id = ?", problem.ID).Delete(&model.ProblemTag{}).Error; err != nil {
		return fmt.Errorf("delete old problem tags: %w", err)
	}

	for _, tagID := range tagIDs {
		problemTag := model.ProblemTag{
			ProblemID: problem.ID,
			TagID:     tagID,
		}
		if err := db.Create(&problemTag).Error; err != nil {
			return fmt.Errorf("create problem tag association: %w", err)
		}
	}

	// Delete existing test cases for this problem and recreate
	if err := db.Where("problem_id = ?", problem.ID).Delete(&model.TestCase{}).Error; err != nil {
		return fmt.Errorf("delete old test cases: %w", err)
	}

	// Seed test cases
	orderIndex := 1

	// Sample test cases
	samplesDir := filepath.Join(problemDir, "samples")
	sampleCases, err := readTestCases(samplesDir, true, &orderIndex)
	if err != nil {
		return fmt.Errorf("read sample test cases: %w", err)
	}

	// Hidden test cases
	hiddenDir := filepath.Join(problemDir, "hidden")
	hiddenCases, err := readTestCases(hiddenDir, false, &orderIndex)
	if err != nil {
		return fmt.Errorf("read hidden test cases: %w", err)
	}

	// Insert all test cases
	allCases := append(sampleCases, hiddenCases...)
	for i := range allCases {
		allCases[i].ProblemID = problem.ID
		if err := db.Create(&allCases[i]).Error; err != nil {
			return fmt.Errorf("create test case %d: %w", allCases[i].OrderIndex, err)
		}
	}

	log.Printf("    Created %d sample + %d hidden test cases", len(sampleCases), len(hiddenCases))

	return nil
}

// ─── Test case reading ──────────────────────────────────────────────────────

// testCaseGroup represents a grouped set of files for a single test case.
type testCaseGroup struct {
	index       int
	inputFile   string
	outputFile  string
	explainFile string // only for samples
}

// readTestCases reads test case files from a directory (samples or hidden).
func readTestCases(dir string, isSample bool, orderIndex *int) ([]model.TestCase, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	// Group files by their base name (e.g., sample1, hidden2)
	groups := make(map[int]*testCaseGroup)
	re := regexp.MustCompile(`^(?:sample|hidden)(\d+)\.(in|out|explanation\.md)$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		idx, _ := strconv.Atoi(matches[1])
		ext := matches[2]

		if groups[idx] == nil {
			groups[idx] = &testCaseGroup{index: idx}
		}

		fullPath := filepath.Join(dir, entry.Name())
		switch ext {
		case "in":
			groups[idx].inputFile = fullPath
		case "out":
			groups[idx].outputFile = fullPath
		case "explanation.md":
			groups[idx].explainFile = fullPath
		}
	}

	// Sort by index
	var indices []int
	for idx := range groups {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	var cases []model.TestCase
	for _, idx := range indices {
		g := groups[idx]
		if g.inputFile == "" || g.outputFile == "" {
			log.Printf("    Warning: skipping incomplete test case group %d in %s", idx, dir)
			continue
		}

		inputData, err := os.ReadFile(g.inputFile)
		if err != nil {
			return nil, fmt.Errorf("read input file %s: %w", g.inputFile, err)
		}

		outputData, err := os.ReadFile(g.outputFile)
		if err != nil {
			return nil, fmt.Errorf("read output file %s: %w", g.outputFile, err)
		}

		var explanation string
		if g.explainFile != "" {
			explainData, err := os.ReadFile(g.explainFile)
			if err != nil {
				return nil, fmt.Errorf("read explanation file %s: %w", g.explainFile, err)
			}
			explanation = string(explainData)
		}

		tc := model.TestCase{
			ID:             uuid.New().String(),
			OrderIndex:     *orderIndex,
			InputData:      string(inputData),
			ExpectedOutput: string(outputData),
			IsSample:       isSample,
			ExplanationMD:  explanation,
		}
		*orderIndex++
		cases = append(cases, tc)
	}

	return cases, nil
}

// ─── Statement parsing ──────────────────────────────────────────────────────

// parseStatementSections extracts sections from the statement markdown.
// It looks for ## Input Format, ## Output Format, ## Constraints, and ## Hints headings.
func parseStatementSections(md string) (inputFormat, outputFormat, constraints, hints string) {
	sections := map[string]*string{
		"输入格式":        &inputFormat,
		"输出格式":        &outputFormat,
		"数据范围":        &constraints,
		"提示":          &hints,
		"input":       &inputFormat,
		"output":      &outputFormat,
		"constraints": &constraints,
		"hints":       &hints,
	}

	lines := strings.Split(md, "\n")
	var currentTarget *string
	var currentLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a heading that matches a section
		if strings.HasPrefix(trimmed, "## ") {
			// Save previous section
			if currentTarget != nil && len(currentLines) > 0 {
				*currentTarget = strings.Join(currentLines, "\n")
			}

			heading := strings.TrimPrefix(trimmed, "## ")
			heading = strings.TrimSpace(heading)
			currentTarget = nil
			currentLines = nil

			// Check for matching section (case-insensitive)
			lowerHeading := strings.ToLower(heading)
			for key, target := range sections {
				if strings.Contains(lowerHeading, strings.ToLower(key)) {
					currentTarget = target
					currentLines = []string{}
					break
				}
			}
			continue
		}

		if currentTarget != nil {
			currentLines = append(currentLines, line)
		}
	}

	// Save last section
	if currentTarget != nil && len(currentLines) > 0 {
		*currentTarget = strings.Join(currentLines, "\n")
	}

	return
}
