package model

import (
	"encoding/json"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
)

// DomainEntryJSON is a wrapper type for plugin.DomainEntry that provides custom JSON marshaling
type DomainEntryJSON struct {
	*pb.DomainEntry
}

// MarshalJSON implements the json.Marshaler interface
func (d DomainEntryJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Domain           string   `json:"domain"`
		AlternativeNames []string `json:"alternativeNames"`
		Alias            string   `json:"alias"`
		Enabled          bool     `json:"enabled"`
		Comment          string   `json:"comment"`
	}{
		Domain:           d.GetDomain(),
		AlternativeNames: d.GetAlternativeNames(),
		Alias:            d.GetAlias(),
		Enabled:          d.GetEnabled(),
		Comment:          d.GetComment(),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *DomainEntryJSON) UnmarshalJSON(data []byte) error {
	var v struct {
		Domain           string   `json:"domain"`
		AlternativeNames []string `json:"alternativeNames"`
		Alias            string   `json:"alias"`
		Enabled          bool     `json:"enabled"`
		Comment          string   `json:"comment"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	d.DomainEntry = &pb.DomainEntry{
		Domain:           v.Domain,
		AlternativeNames: v.AlternativeNames,
		Alias:            v.Alias,
		Enabled:          v.Enabled,
		Comment:          v.Comment,
	}
	return nil
}

// NewDomainEntryJSON creates a new DomainEntryJSON wrapper
func NewDomainEntryJSON(d *pb.DomainEntry) *DomainEntryJSON {
	return &DomainEntryJSON{DomainEntry: d}
}
