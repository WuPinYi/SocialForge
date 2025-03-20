package schema

import "entgo.io/ent"

// Influencer holds the schema definition for the Influencer entity.
type Influencer struct {
	ent.Schema
}

// Fields of the Influencer.
func (Influencer) Fields() []ent.Field {
	return nil
}

// Edges of the Influencer.
func (Influencer) Edges() []ent.Edge {
	return nil
}
