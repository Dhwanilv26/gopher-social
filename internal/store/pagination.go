package store

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int        `json:"limit" validate:"gte=1,lte=20"`
	Offset int        `json:"offset" validate:"gte=0"`
	Sort   string     `json:"sort" validate:"oneof=asc desc"`
	Tags   []string   `json:"tags" validate:"max=5"`
	Search string     `json:"search" validate:"max=100"` // can filter both title and content here using the same search keyword
	Since  *time.Time `json:"since"`
	Until  *time.Time `json:"until"`
}

func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, err
		}
		fq.Limit = l
	}

	offset := qs.Get("offset")

	if offset != "" {
		l, err := strconv.Atoi(offset)
		if err != nil {
			return fq, err
		}
		fq.Offset = l
	}

	sort := qs.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	tags := qs.Get("tags")
	if tags != "" {
		for _, tag := range strings.Split(tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				fq.Tags = append(fq.Tags, tag)
			}
		}
	}

	search := qs.Get("search")
	if search != "" {
		fq.Search = search
	}

	since := qs.Get("since")
	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			return fq, fmt.Errorf("invalid since timestamp: %w", err)
		}
		fq.Since = &t
	}

	until := qs.Get("until")
	if until != "" {
		t, err := time.Parse(time.RFC3339, until)
		if err != nil {
			return fq, fmt.Errorf("invalid until timestamp: %w", err)
		}
		fq.Until = &t
	}

	if fq.Since != nil && fq.Until != nil &&
		fq.Since.After(*fq.Until) {
		return fq, fmt.Errorf("since must not be after until")
	}

	return fq, nil
}
