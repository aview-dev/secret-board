package board

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "board.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreatedAtIsBlurredToHour(t *testing.T) {
	s := openTest(t)
	s.now = func() time.Time { return time.Date(2026, 10, 21, 14, 37, 12, 0, time.UTC) }
	want := time.Date(2026, 10, 21, 14, 0, 0, 0, time.UTC)

	p, err := s.CreatePost(context.Background(), "제목", "본문")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetPost(context.Background(), p.ID)
	if !got.CreatedAt.Equal(want) {
		t.Fatalf("post createdAt: got %v, want %v", got.CreatedAt, want)
	}

	c, err := s.CreateComment(context.Background(), p.ID, "댓글")
	if err != nil {
		t.Fatal(err)
	}
	if !c.CreatedAt.Equal(want) {
		t.Fatalf("comment createdAt: got %v, want %v", c.CreatedAt, want)
	}
}

func TestPostAndComments(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()

	p1, _ := s.CreatePost(ctx, "첫 글", "a")
	p2, _ := s.CreatePost(ctx, "  둘째 글  ", "b")
	if p2.Title != "둘째 글" {
		t.Fatalf("title not trimmed: %q", p2.Title)
	}
	s.CreateComment(ctx, p1.ID, "1")
	s.CreateComment(ctx, p1.ID, "2")

	posts, err := s.ListPosts(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 || posts[0].ID != p2.ID || posts[1].CommentCount != 2 {
		t.Fatalf("unexpected list: %+v", posts)
	}

	comments, _ := s.ListComments(ctx, p1.ID)
	if len(comments) != 2 || comments[0].Body != "1" {
		t.Fatalf("unexpected comments: %+v", comments)
	}
}

func TestValidation(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()

	if _, err := s.CreatePost(ctx, "   ", "본문"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("blank title: got %v", err)
	}
	if _, err := s.CreatePost(ctx, strings.Repeat("가", MaxTitleLen+1), "본문"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("long title: got %v", err)
	}
	if _, err := s.CreateComment(ctx, 999, "댓글"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("comment on missing post: got %v", err)
	}
	if _, err := s.GetPost(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing post: got %v", err)
	}
}
