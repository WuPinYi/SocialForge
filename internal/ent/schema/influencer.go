//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate .

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Influencer holds the schema definition for the Influencer entity.
type Influencer struct {
	ent.Schema
}

// Fields of the Influencer.
func (Influencer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("name"),
		field.String("platform"),
		field.String("account_id"),
		field.String("status").
			Default("active"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Influencer.
func (Influencer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("influencers").
			Unique().
			Required(),
		edge.To("posts", Post.Type),
	}
}

// Indexes of the Influencer.
func (Influencer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "account_id").Unique(),
	}
}
