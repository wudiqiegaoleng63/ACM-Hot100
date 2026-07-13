package repository

import (
	"github.com/acmhot100/server/internal/model"
	"gorm.io/gorm"
)

// ProblemWithStatus extends Problem with the user's progress state.
type ProblemWithStatus struct {
	model.Problem
	State string `json:"state"` // NOT_STARTED, ATTEMPTED, SOLVED, or ""
}

// ProblemNav holds minimal info for previous/next navigation.
type ProblemNav struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

// ListProblems returns a paginated list of published problems with optional filters.
// If userID is non-empty, each problem includes the user's progress state.
func ListProblems(db *gorm.DB, q, difficulty, tag, state, userID string, page, pageSize int) ([]ProblemWithStatus, int, error) {
	base := db.Model(&model.Problem{}).Where("is_published = ?", true)

	// Search by title
	if q != "" {
		base = base.Where("title LIKE ?", "%"+q+"%")
	}
	// Filter by difficulty
	if difficulty != "" {
		base = base.Where("difficulty = ?", difficulty)
	}
	// Filter by tag
	if tag != "" {
		base = base.Joins("JOIN problem_tags ON problem_tags.problem_id = problems.id").
			Joins("JOIN tags ON tags.id = problem_tags.tag_id").
			Where("tags.slug = ?", tag)
	}

	// Count total matching rows
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch problems ordered by order_index with pagination
	var problems []model.Problem
	offset := (page - 1) * pageSize
	if err := base.Preload("Tags").
		Order("order_index ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	// Build result with optional user state
	results := make([]ProblemWithStatus, len(problems))

	if userID != "" && len(problems) > 0 {
		// Collect problem IDs
		problemIDs := make([]string, len(problems))
		for i, p := range problems {
			problemIDs[i] = p.ID
		}

		// Fetch progress for all problems in one query
		var progress []model.UserProblemProgress
		if err := db.Where("user_id = ? AND problem_id IN ?", userID, problemIDs).Find(&progress).Error; err != nil {
			return nil, 0, err
		}

		// Build a map of problemID -> state
		progressMap := make(map[string]string, len(progress))
		for _, p := range progress {
			progressMap[p.ProblemID] = p.State
		}

		// Filter by state if requested
		for i, p := range problems {
			s, ok := progressMap[p.ID]
			if !ok {
				s = model.ProgressNotStarted
			}
			if state != "" && s != state {
				continue
			}
			results[i] = ProblemWithStatus{Problem: p, State: s}
		}

		// If state filter was applied, we need to re-filter and re-count
		if state != "" {
			var filtered []ProblemWithStatus
			for _, r := range results {
				if r.State == state {
					filtered = append(filtered, r)
				}
			}
			// Recalculate total for state filter
			// We need a separate count query that joins progress
			countBase := db.Model(&model.Problem{}).Where("is_published = ?", true)
			if q != "" {
				countBase = countBase.Where("title LIKE ?", "%"+q+"%")
			}
			if difficulty != "" {
				countBase = countBase.Where("difficulty = ?", difficulty)
			}
			if tag != "" {
				countBase = countBase.Joins("JOIN problem_tags ON problem_tags.problem_id = problems.id").
					Joins("JOIN tags ON tags.id = problem_tags.tag_id").
					Where("tags.slug = ?", tag)
			}
			countBase = countBase.Joins("LEFT JOIN user_problem_progress upp ON upp.problem_id = problems.id AND upp.user_id = ?", userID)
			if state == model.ProgressNotStarted {
				countBase = countBase.Where("upp.state IS NULL")
			} else {
				countBase = countBase.Where("upp.state = ?", state)
			}
			if err := countBase.Count(&total).Error; err != nil {
				return nil, 0, err
			}

			// Paginate filtered results
			start := offset
			end := start + pageSize
			if start > len(filtered) {
				start = len(filtered)
			}
			if end > len(filtered) {
				end = len(filtered)
			}
			return filtered[start:end], int(total), nil
		}
	} else if state != "" {
		// State filter requested but no user -> return empty
		return []ProblemWithStatus{}, 0, nil
	}

	// No state filter or no userID: just map problems to results
	for i, p := range problems {
		results[i] = ProblemWithStatus{Problem: p}
	}

	return results, int(total), nil
}

// GetProblemBySlug returns a single problem by its slug, or nil if not found.
func GetProblemBySlug(db *gorm.DB, slug string) (*model.Problem, error) {
	var problem model.Problem
	if err := db.Preload("Tags").Where("slug = ?", slug).First(&problem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &problem, nil
}

// GetProblemNavigation returns the previous and next problem by order_index.
func GetProblemNavigation(db *gorm.DB, slug string) (*ProblemNav, *ProblemNav, error) {
	// First get the current problem's order_index
	var current model.Problem
	if err := db.Where("slug = ? AND is_published = ?", slug, true).First(&current).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var prevNav *ProblemNav
	var nextNav *ProblemNav

	// Previous: largest order_index less than current
	var prevProblem model.Problem
	if err := db.Where("is_published = ? AND order_index < ?", true, current.OrderIndex).
		Order("order_index DESC").First(&prevProblem).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, nil, err
		}
		prevNav = nil
	} else {
		prevNav = &ProblemNav{Slug: prevProblem.Slug, Title: prevProblem.Title}
	}

	// Next: smallest order_index greater than current
	var nextProblem model.Problem
	if err := db.Where("is_published = ? AND order_index > ?", true, current.OrderIndex).
		Order("order_index ASC").First(&nextProblem).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, nil, err
		}
		nextNav = nil
	} else {
		nextNav = &ProblemNav{Slug: nextProblem.Slug, Title: nextProblem.Title}
	}

	return prevNav, nextNav, nil
}
