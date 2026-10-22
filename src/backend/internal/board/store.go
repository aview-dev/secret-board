// 게시글 / 댓글 저장소 (board.db).
//
// 익명성 규칙 (docs/anonymity.md):
//   - 규칙 1: 어떤 테이블에도 작성자 식별자 컬럼을 두지 않는다. 글 내용과 뭉갠 시각만 저장한다.
//   - 규칙 4: 작성 시각은 저장할 때 시간 단위로 내림한다. 정확한 시각은 DB 어디에도 남지 않는다.
//   - 규칙 9: 투표는 portal 이 만든 글마다 다른 해시(voter)로만 구분한다. 투표 시각은 저장하지 않는다.

package board

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite" // 순수 Go SQLite 드라이버 (CGO 없이 정적 바이너리로 빌드하기 위함)
)

// 글자 수 제한 (바이트가 아니라 글자 기준).
const (
	MaxTitleLen   = 100
	MaxBodyLen    = 10000
	MaxCommentLen = 2000
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
)

type Post struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Body         string    `json:"body,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	CommentCount int       `json:"commentCount"`
	Likes        int       `json:"likes"`
	Dislikes     int       `json:"dislikes"`
}

// Votes 는 투표 뒤 화면에 다시 그릴 값이다. Mine 은 요청한 voter 의 표 (1, -1, 0).
type Votes struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
	Mine     int `json:"mine"`
}

// 신고 사유. 자유 입력은 받지 않는다 (사유 글에 신원이 드러날 수 있어서). schema 의 CHECK 와 맞춘다.
var ReportReasons = []string{"spam", "abuse", "privacy", "other"}

// ReportedPost 는 관리자 신고 목록의 한 줄이다. 누가 신고했는지는 없고 개수만 있다.
type ReportedPost struct {
	ID        int64          `json:"id"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"createdAt"`
	Reports   int            `json:"reports"`
	Reasons   map[string]int `json:"reasons"`
}

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"postId"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	db *sql.DB
	// 테스트에서 시각을 고정하려고 바꿔 끼운다.
	now func() time.Time
}

const schema = `
CREATE TABLE IF NOT EXISTS posts (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	title      TEXT    NOT NULL,
	body       TEXT    NOT NULL,
	created_at INTEGER NOT NULL -- unix 초, 시간 단위로 내림
);
CREATE TABLE IF NOT EXISTS comments (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	post_id    INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	body       TEXT    NOT NULL,
	created_at INTEGER NOT NULL -- unix 초, 시간 단위로 내림
);
CREATE INDEX IF NOT EXISTS comments_post_id ON comments(post_id, id);
-- voter 는 HMAC(portal 비밀키, userId:postId) 의 hex. 글마다 달라서 한 사람의 표끼리 묶을 수 없다.
-- WITHOUT ROWID: 행이 (post_id, voter) 순서로만 저장되어 누가 먼저 눌렀는지 순서가 남지 않는다.
CREATE TABLE IF NOT EXISTS votes (
	post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	voter   TEXT    NOT NULL,
	value   INTEGER NOT NULL CHECK (value IN (1, -1)),
	PRIMARY KEY (post_id, voter)
) WITHOUT ROWID;
-- reporter 는 HMAC(신고 키, userId:postId) 의 hex. 투표와 키가 달라 voter 와 맞춰 볼 수 없다 (규칙 11).
-- 신고 시각은 두지 않고, WITHOUT ROWID 라 신고 순서도 남지 않는다.
CREATE TABLE IF NOT EXISTS reports (
	post_id  INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	reporter TEXT    NOT NULL,
	reason   TEXT    NOT NULL CHECK (reason IN ('spam', 'abuse', 'privacy', 'other')),
	PRIMARY KEY (post_id, reporter)
) WITHOUT ROWID;
`

// 글 하나의 좋아요 / 싫어요 수. posts 를 p 로 부르는 쿼리 안에서 쓴다.
const voteCounts = `
	(SELECT COUNT(*) FROM votes v WHERE v.post_id = p.id AND v.value = 1),
	(SELECT COUNT(*) FROM votes v WHERE v.post_id = p.id AND v.value = -1)`

// Open 은 path 의 SQLite 파일을 열고 (없으면 만들고) 스키마를 맞춘다.
func Open(path string) (*Store, error) {
	// foreign_keys 는 연결마다 켜야 해서 DSN 에 둔다. busy_timeout 은 쓰기가 겹칠 때 바로 실패하지 않게 한다.
	// secure_delete: 지운 행이 있던 자리를 0 으로 덮어쓴다. 끄면 지운 글이 파일의 빈 페이지에 한동안 남아
	// board.db 를 열어 보면 되살릴 수 있다 (관리자 삭제의 의미가 없어진다).
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=secure_delete(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite 는 쓰기가 한 번에 하나뿐이라 연결 하나로 충분하고, 잠금 충돌도 피한다.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db, now: time.Now}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// blurredNow 는 현재 시각을 시간 단위로 내린 unix 초다 (규칙 4).
func (s *Store) blurredNow() int64 {
	return s.now().UTC().Truncate(time.Hour).Unix()
}

// clean 은 앞뒤 공백을 지우고 1 글자 이상, max 글자 이하인지 확인한다.
func clean(v string, max int) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" || !utf8.ValidString(v) || utf8.RuneCountInString(v) > max {
		return "", ErrInvalid
	}
	return v, nil
}

func (s *Store) CreatePost(ctx context.Context, title, body string) (Post, error) {
	title, err := clean(title, MaxTitleLen)
	if err != nil {
		return Post{}, err
	}
	body, err = clean(body, MaxBodyLen)
	if err != nil {
		return Post{}, err
	}
	at := s.blurredNow()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO posts (title, body, created_at) VALUES (?, ?, ?)`, title, body, at)
	if err != nil {
		return Post{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Post{}, err
	}
	return Post{ID: id, Title: title, Body: body, CreatedAt: time.Unix(at, 0).UTC()}, nil
}

// ListPosts 는 최신 글부터 limit 개를 돌려준다. 본문은 빼고 댓글 수를 붙인다.
func (s *Store) ListPosts(ctx context.Context, limit int) ([]Post, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.created_at, (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id),`+voteCounts+`
		FROM posts p ORDER BY p.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var p Post
		var at int64
		if err := rows.Scan(&p.ID, &p.Title, &at, &p.CommentCount, &p.Likes, &p.Dislikes); err != nil {
			return nil, err
		}
		p.CreatedAt = time.Unix(at, 0).UTC()
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (s *Store) GetPost(ctx context.Context, id int64) (Post, error) {
	var p Post
	var at int64
	err := s.db.QueryRowContext(ctx, `
		SELECT p.id, p.title, p.body, p.created_at, (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id),`+voteCounts+`
		FROM posts p WHERE p.id = ?`, id).Scan(&p.ID, &p.Title, &p.Body, &at, &p.CommentCount, &p.Likes, &p.Dislikes)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if err != nil {
		return Post{}, err
	}
	p.CreatedAt = time.Unix(at, 0).UTC()
	return p, nil
}

// DeletePost 는 글을 지운다. 댓글과 표는 ON DELETE CASCADE 로 함께 지워진다.
// 누가 지웠는지는 받지도 남기지도 않는다 (관리자 확인은 portal 이 한다).
func (s *Store) DeletePost(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteComment 는 글 postID 에 달린 댓글 하나를 지운다. 다른 글의 댓글 id 를 넘기면 ErrNotFound.
func (s *Store) DeleteComment(ctx context.Context, postID, commentID int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM comments WHERE id = ? AND post_id = ?`, commentID, postID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListComments 는 오래된 댓글부터 돌려준다.
func (s *Store) ListComments(ctx context.Context, postID int64) ([]Comment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, post_id, body, created_at FROM comments WHERE post_id = ? ORDER BY id`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var c Comment
		var at int64
		if err := rows.Scan(&c.ID, &c.PostID, &c.Body, &at); err != nil {
			return nil, err
		}
		c.CreatedAt = time.Unix(at, 0).UTC()
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (s *Store) CreateComment(ctx context.Context, postID int64, body string) (Comment, error) {
	body, err := clean(body, MaxCommentLen)
	if err != nil {
		return Comment{}, err
	}
	at := s.blurredNow()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO comments (post_id, body, created_at)
		 SELECT id, ?, ? FROM posts WHERE id = ?`, body, at, postID)
	if err != nil {
		return Comment{}, err
	}
	// 글이 없으면 SELECT 가 비어서 아무것도 들어가지 않는다.
	if n, _ := res.RowsAffected(); n == 0 {
		return Comment{}, ErrNotFound
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Comment{}, err
	}
	return Comment{ID: id, PostID: postID, Body: body, CreatedAt: time.Unix(at, 0).UTC()}, nil
}

// validHash 는 voter / reporter 가 SHA-256 hex (소문자 64 글자) 인지 본다.
// portal 이 실수로 이메일 같은 원래 값을 보내도 저장되지 않게 하는 마지막 방어선이다.
func validHash(v string) bool {
	if len(v) != 64 {
		return false
	}
	for _, c := range v {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// MyVote 는 voter 가 글에 준 표를 돌려준다. 안 눌렀거나 voter 가 비어 있으면 0.
func (s *Store) MyVote(ctx context.Context, postID int64, voter string) (int, error) {
	if !validHash(voter) {
		return 0, nil
	}
	var v int
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM votes WHERE post_id = ? AND voter = ?`, postID, voter).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return v, err
}

// Vote 는 voter 의 표를 value 로 바꾼다. 1 = 좋아요, -1 = 싫어요, 0 = 취소.
// 같은 voter 는 글마다 한 표뿐이라, 좋아요를 누른 채 싫어요를 누르면 싫어요로 바뀐다.
func (s *Store) Vote(ctx context.Context, postID int64, voter string, value int) (Votes, error) {
	if !validHash(voter) || value < -1 || value > 1 {
		return Votes{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Votes{}, err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM posts WHERE id = ?)`, postID).Scan(&exists); err != nil {
		return Votes{}, err
	}
	if !exists {
		return Votes{}, ErrNotFound
	}

	if value == 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM votes WHERE post_id = ? AND voter = ?`, postID, voter)
	} else {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO votes (post_id, voter, value) VALUES (?, ?, ?)
			ON CONFLICT (post_id, voter) DO UPDATE SET value = excluded.value`, postID, voter, value)
	}
	if err != nil {
		return Votes{}, err
	}

	out := Votes{Mine: value}
	if err := tx.QueryRowContext(ctx, `SELECT`+voteCounts+` FROM posts p WHERE p.id = ?`, postID).
		Scan(&out.Likes, &out.Dislikes); err != nil {
		return Votes{}, err
	}
	return out, tx.Commit()
}

func validReason(r string) bool {
	for _, v := range ReportReasons {
		if r == v {
			return true
		}
	}
	return false
}

// Report 는 reporter 의 신고를 남긴다. 같은 reporter 가 다시 신고하면 사유만 바뀌고 개수는 그대로다.
func (s *Store) Report(ctx context.Context, postID int64, reporter, reason string) error {
	if !validHash(reporter) || !validReason(reason) {
		return ErrInvalid
	}
	// 글이 없으면 SELECT 가 비어서 아무것도 들어가지 않는다 (CreateComment 와 같은 방식).
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO reports (post_id, reporter, reason)
		SELECT id, ?, ? FROM posts WHERE id = ?
		ON CONFLICT (post_id, reporter) DO UPDATE SET reason = excluded.reason`, reporter, reason, postID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListReported 는 신고가 하나라도 있는 글을 신고가 많은 순으로 돌려준다.
func (s *Store) ListReported(ctx context.Context) ([]ReportedPost, error) {
	// 사유별로 한 줄씩 나온다. total 과 id 로 정렬하므로 같은 글의 줄은 붙어서 나온다.
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.created_at, r.reason, COUNT(*),
			(SELECT COUNT(*) FROM reports t WHERE t.post_id = p.id) AS total
		FROM reports r JOIN posts p ON p.id = r.post_id
		GROUP BY p.id, r.reason
		ORDER BY total DESC, p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ReportedPost{}
	for rows.Next() {
		var (
			id           int64
			title        string
			at           int64
			reason       string
			count, total int
		)
		if err := rows.Scan(&id, &title, &at, &reason, &count, &total); err != nil {
			return nil, err
		}
		if len(out) == 0 || out[len(out)-1].ID != id {
			out = append(out, ReportedPost{
				ID: id, Title: title, CreatedAt: time.Unix(at, 0).UTC(), Reports: total, Reasons: map[string]int{},
			})
		}
		out[len(out)-1].Reasons[reason] = count
	}
	return out, rows.Err()
}

// ClearReports 는 관리자가 "문제없음" 으로 판단한 글의 신고를 모두 지운다. 이후 같은 사람이 다시 신고할 수 있다.
func (s *Store) ClearReports(ctx context.Context, postID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM reports WHERE post_id = ?`, postID)
	return err
}
