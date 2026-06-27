package site

import (
	"fmt"
	"hachigo/pkg/content"
)

// Paginator represents the paginator object exposed to index templates
type Paginator struct {
	Posts            []map[string]interface{} `json:"posts"`
	Page             int                      `json:"page"`
	PerPage          int                      `json:"per_page"`
	TotalPosts       int                      `json:"total_posts"`
	TotalPages       int                      `json:"total_pages"`
	PreviousPage     int                      `json:"previous_page"` // 0 if none
	PreviousPagePath string                   `json:"previous_page_path"`
	NextPage         int                      `json:"next_page"` // 0 if none
	NextPagePath     string                   `json:"next_page_path"`
}

// ToMap converts the Paginator to a map representation for Liquid
func (p *Paginator) ToMap() map[string]interface{} {
	var prevPageVal interface{}
	if p.PreviousPage > 0 {
		prevPageVal = p.PreviousPage
	}
	var nextPageVal interface{}
	if p.NextPage > 0 {
		nextPageVal = p.NextPage
	}

	return map[string]interface{}{
		"posts":              p.Posts,
		"page":               p.Page,
		"per_page":           p.PerPage,
		"total_posts":        p.TotalPosts,
		"total_pages":        p.TotalPages,
		"previous_page":      prevPageVal,
		"previous_page_path": p.PreviousPagePath,
		"next_page":          nextPageVal,
		"next_page_path":     p.NextPagePath,
	}
}

// PaginatePosts slices sorted posts into paginated groups
func PaginatePosts(posts []*content.Post, postsPerPage int) []Paginator {
	if postsPerPage <= 0 {
		postsPerPage = 10
	}

	totalPosts := len(posts)
	totalPages := (totalPosts + postsPerPage - 1) / postsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	paginators := make([]Paginator, totalPages)
	for i := 0; i < totalPages; i++ {
		start := i * postsPerPage
		end := start + postsPerPage
		if end > totalPosts {
			end = totalPosts
		}

		pagePosts := posts[start:end]
		pageMaps := make([]map[string]interface{}, len(pagePosts))
		for j, p := range pagePosts {
			pageMaps[j] = p.ToMap()
		}

		pageNum := i + 1
		var prevPage, nextPage int
		var prevPath, nextPath string

		if pageNum > 1 {
			prevPage = pageNum - 1
			if prevPage == 1 {
				prevPath = "/"
			} else {
				prevPath = fmt.Sprintf("/posts/%d/", prevPage)
			}
		}

		if pageNum < totalPages {
			nextPage = pageNum + 1
			nextPath = fmt.Sprintf("/posts/%d/", nextPage)
		}

		paginators[i] = Paginator{
			Posts:            pageMaps,
			Page:             pageNum,
			PerPage:          postsPerPage,
			TotalPosts:       totalPosts,
			TotalPages:       totalPages,
			PreviousPage:     prevPage,
			PreviousPagePath: prevPath,
			NextPage:         nextPage,
			NextPagePath:     nextPath,
		}
	}

	return paginators
}
