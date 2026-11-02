package catalog

// Record is one Application's Catalog change as aiport keeps it: which
// configuration Change made it and when. A revert names the Change it undid
// in RevertOf. It holds what Change holds (Application ids, task names,
// model names and Connection ids) and Change ids, nothing else.
type Record struct {
	ChangeID string `json:"change_id"`
	At       int64  `json:"at_ms"`
	RevertOf string `json:"revert_of,omitempty"`
	Change
}
