package models

// BinType mirrors the Postgres `bin_type` enum. Kept as a plain string type
// (not iota) so it serializes to JSON exactly as stored in the database.
type BinType string

const (
	BinOrganic   BinType = "organic"
	BinAnorganic BinType = "anorganic"
)

func (b BinType) Valid() bool {
	return b == BinOrganic || b == BinAnorganic
}

// NormalizeBinType lowercases and matches common casing variants a device
// might send (e.g. "Organic", "ANORGANIC"), returning ok=false if it still
// doesn't match a known bin type after normalization.
func NormalizeBinType(raw string) (BinType, bool) {
	switch normalizeToken(raw) {
	case "organic":
		return BinOrganic, true
	case "anorganic", "inorganic", "non-organic", "nonorganic":
		return BinAnorganic, true
	default:
		return "", false
	}
}
