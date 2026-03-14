package query

// UserFilter allows filtering user queries by various criteria.
type UserFilter struct {
	NameLike *string
	Limit    int
	Offset   int
	OrderBy  string // "asc" or "desc" (default: "asc")
	SortBy   string // "name" or "created_at" (default: "name")
}
