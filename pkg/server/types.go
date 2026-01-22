package server

// SearchArgs represents arguments for the search tool
type SearchArgs struct {
	Query        string   `json:"query" jsonschema:"Search query"`
	TopK         int      `json:"top_k,omitempty" jsonschema:"Number of results to return (default: 5)"`
	MinScore     *float64 `json:"min_score,omitempty" jsonschema:"Minimum similarity score 0-1 (default: 0.3)"`
	SourceFilter string   `json:"source_filter,omitempty" jsonschema:"Filter results to specific source (file path or URL)"`
}

// IndexArgs represents arguments for the index tool
type IndexArgs struct {
	FilePath string `json:"file_path,omitempty" jsonschema:"Path to file to index"`
	URL      string `json:"url,omitempty" jsonschema:"URL to fetch and index"`
	Content  string `json:"content,omitempty" jsonschema:"Direct content to index (requires source)"`
	Source   string `json:"source,omitempty" jsonschema:"Source identifier when using content parameter"`
	Reindex  bool   `json:"reindex,omitempty" jsonschema:"Force re-index if already indexed (default: false)"`
}

// ListArgs represents arguments for the list tool
type ListArgs struct {
	SourceType string `json:"source_type,omitempty" jsonschema:"Filter by source type: 'file' or 'url' (empty for all)"`
}

// DeleteArgs represents arguments for the delete tool
type DeleteArgs struct {
	Source string `json:"source" jsonschema:"Source to delete (file path or URL)"`
}
