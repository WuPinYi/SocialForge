//go:generate go run -mod=mod entgo.io/ent/cmd/ent init Post

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("influencer_id"),
		field.Text("content"),
		field.Time("scheduled_time"),
		field.String("status").
			Default("scheduled"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("influencer", Influencer.Type).
			Ref("posts").
			Field("influencer_id").
			Unique().
			Required(),
	}
}

// Indexes of the Post.
func (Post) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("influencer_id", "scheduled_time"),
		index.Fields("status"),
	}
}
