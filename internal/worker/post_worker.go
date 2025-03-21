package worker

import (
	"context"
	"log"
	"time"

	"github.com/WuPinYi/SocialForge/internal/ent"
	"github.com/WuPinYi/SocialForge/internal/ent/post"
)

type PostWorker struct {
	client *ent.Client
}

func NewPostWorker(client *ent.Client) *PostWorker {
	return &PostWorker{
		client: client,
	}
}

func (w *PostWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processScheduledPosts(ctx); err != nil {
				log.Printf("Error processing scheduled posts: %v", err)
			}
		}
	}
}

func (w *PostWorker) processScheduledPosts(ctx context.Context) error {
	// Find all posts that are scheduled and due
	posts, err := w.client.Post.Query().
		Where(
			post.Status("scheduled"),
			post.ScheduledTimeLTE(time.Now()),
		).
		All(ctx)
	if err != nil {
		return err
	}

	for _, p := range posts {
		// Get the influencer for this post
		influencer, err := p.QueryInfluencer().Only(ctx)
		if err != nil {
			log.Printf("Error getting influencer for post %s: %v", p.ID, err)
			continue
		}

		// TODO: Implement actual posting logic here
		// This would involve:
		// 1. Getting the appropriate social media client
		// 2. Posting the content
		// 3. Updating the post status

		// For now, we'll just update the status
		_, err = p.Update().
			SetStatus("posted").
			Save(ctx)
		if err != nil {
			log.Printf("Error updating post status %s: %v", p.ID, err)
			continue
		}

		log.Printf("Successfully processed post %s for influencer %s", p.ID, influencer.Name)
	}

	return nil
}
