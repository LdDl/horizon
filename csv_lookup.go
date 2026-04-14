package horizon

import (
	"fmt"
	"strconv"
)

// CSVLookup provides column name-based access to CSV records.
// Instead of positional indexing (record[0], record[1], ...) it allows
// accessing fields by header name: lookup.Int64(record, "source_vertex_id").
type CSVLookup struct {
	columns map[string]int
}

// NewCSVLookup creates a CSVLookup from a CSV header row.
func NewCSVLookup(header []string) *CSVLookup {
	columns := make(map[string]int, len(header))
	for i, name := range header {
		columns[name] = i
	}
	return &CSVLookup{columns: columns}
}

// Has returns true if the column exists in the header.
func (l *CSVLookup) Has(name string) bool {
	_, ok := l.columns[name]
	return ok
}

// Index returns the column index for the given name.
func (l *CSVLookup) Index(name string) (int, error) {
	idx, ok := l.columns[name]
	if !ok {
		return -1, fmt.Errorf("column %q not found in CSV header", name)
	}
	return idx, nil
}

// String returns the string value of the named column.
func (l *CSVLookup) String(record []string, name string) (string, error) {
	idx, err := l.Index(name)
	if err != nil {
		return "", err
	}
	if idx >= len(record) {
		return "", fmt.Errorf("column %q index %d out of range (record has %d fields)", name, idx, len(record))
	}
	return record[idx], nil
}

// Int64 parses the named column as int64.
func (l *CSVLookup) Int64(record []string, name string) (int64, error) {
	s, err := l.String(record, name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("column %q: can't parse %q as int64: %w", name, s, err)
	}
	return v, nil
}

// Int parses the named column as int.
func (l *CSVLookup) Int(record []string, name string) (int, error) {
	s, err := l.String(record, name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("column %q: can't parse %q as int: %w", name, s, err)
	}
	return v, nil
}

// Float64 parses the named column as float64.
func (l *CSVLookup) Float64(record []string, name string) (float64, error) {
	s, err := l.String(record, name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("column %q: can't parse %q as float64: %w", name, s, err)
	}
	return v, nil
}

// MustHave validates that all required columns are present in the header.
// Returns an error listing all missing columns.
func (l *CSVLookup) MustHave(names ...string) error {
	var missing []string
	for _, name := range names {
		if !l.Has(name) {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required CSV columns missing: %v", missing)
	}
	return nil
}
