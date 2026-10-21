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

func TestVote(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	p, _ := s.CreatePost(ctx, "글", "본문")
	alice := strings.Repeat("a", 64)
	bob := strings.Repeat("b", 64)

	check := func(got Votes, likes, dislikes, mine int) {
		t.Helper()
		if got.Likes != likes || got.Dislikes != dislikes || got.Mine != mine {
			t.Fatalf("got %+v, want likes=%d dislikes=%d mine=%d", got, likes, dislikes, mine)
		}
	}

	v, err := s.Vote(ctx, p.ID, alice, 1)
	if err != nil {
		t.Fatal(err)
	}
	check(v, 1, 0, 1)

	// 같은 사람이 다시 눌러도 한 표뿐이다.
	v, _ = s.Vote(ctx, p.ID, alice, 1)
	check(v, 1, 0, 1)

	// 좋아요에서 싫어요로 바꾸기.
	v, _ = s.Vote(ctx, p.ID, alice, -1)
	check(v, 0, 1, -1)

	v, _ = s.Vote(ctx, p.ID, bob, 1)
	check(v, 1, 1, 1)

	// 취소.
	v, _ = s.Vote(ctx, p.ID, alice, 0)
	check(v, 1, 0, 0)

	got, _ := s.GetPost(ctx, p.ID)
	if got.Likes != 1 || got.Dislikes != 0 {
		t.Fatalf("GetPost counts: %+v", got)
	}
	posts, _ := s.ListPosts(ctx, 10)
	if posts[0].Likes != 1 {
		t.Fatalf("ListPosts likes: %+v", posts[0])
	}
	if mine, _ := s.MyVote(ctx, p.ID, bob); mine != 1 {
		t.Fatalf("MyVote bob: got %d", mine)
	}
	if mine, _ := s.MyVote(ctx, p.ID, alice); mine != 0 {
		t.Fatalf("MyVote alice: got %d", mine)
	}
}

func TestVoteRejectsBadInput(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	p, _ := s.CreatePost(ctx, "글", "본문")
	ok := strings.Repeat("a", 64)

	// voter 는 해시만 받는다. 이메일 같은 원래 값은 거절한다.
	for _, voter := range []string{"", "someone@corelinesoft.com", strings.Repeat("A", 64), strings.Repeat("a", 63)} {
		if _, err := s.Vote(ctx, p.ID, voter, 1); !errors.Is(err, ErrInvalid) {
			t.Fatalf("voter %q: got %v, want ErrInvalid", voter, err)
		}
	}
	if _, err := s.Vote(ctx, p.ID, ok, 2); !errors.Is(err, ErrInvalid) {
		t.Fatalf("value 2: got %v", err)
	}
	if _, err := s.Vote(ctx, 9999, ok, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing post: got %v", err)
	}
}
