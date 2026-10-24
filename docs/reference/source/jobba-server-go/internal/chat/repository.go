package chat

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct{ p *pgxpool.Pool }

func NewRepository(p *pgxpool.Pool) Repository { return &repository{p} }

type scan interface{ Scan(...any) error }

const convCols = `c.id::text,c.job_id::text,j.job_title,j.status::text,c.participant_a::text,ua.first_name,ua.last_name,ua.email,ua.account_type,c.participant_b::text,ub.first_name,ub.last_name,ub.email,ub.account_type,c.last_message_at,(cm.archived_at IS NOT NULL),c.is_active,c.is_urgent,c.marked_urgent_by::text,cm.user_id::text,cm.unread_count,c.created_at,c.updated_at`

func scanConv(s scan) (Conversation, error) {
	var c Conversation
	var a, b User
	var mark *string
	var memberID string
	var unread int
	err := s.Scan(&c.ID, &c.Job.ID, &c.Job.JobTitle, &c.Job.Status, &a.ID, &a.FirstName, &a.LastName, &a.Email, &a.AccountType, &b.ID, &b.FirstName, &b.LastName, &b.Email, &b.AccountType, &c.LastMessageAt, &c.IsArchived, &c.IsActive, &c.IsUrgent, &mark, &memberID, &unread, &c.CreatedAt, &c.UpdatedAt)
	c.Participants = []User{a, b}
	c.MarkedUrgentBy = mark
	c.UnreadCount = map[string]int{}
	if a.ID != "" {
		c.UnreadCount[a.ID] = 0
		c.UnreadCount[b.ID] = 0
		c.UnreadCount[memberID] = unread
	}
	return c, err
}

const joins = ` FROM conversations c JOIN jobs j ON j.id=c.job_id JOIN users ua ON ua.id=c.participant_a JOIN users ub ON ub.id=c.participant_b JOIN conversation_members cm ON cm.conversation_id=c.id `

func (r *repository) Create(ctx context.Context, u, p, j string) (Conversation, error) {
	tx, e := r.p.Begin(ctx)
	if e != nil {
		return Conversation{}, e
	}
	defer tx.Rollback(ctx)
	var owner string
	e = tx.QueryRow(ctx, `SELECT client_id::text FROM jobs WHERE id=$1::uuid AND NOT is_deleted`, j).Scan(&owner)
	if errors.Is(e, pgx.ErrNoRows) {
		return Conversation{}, ErrParticipant
	}
	if e != nil {
		return Conversation{}, e
	}
	if u != owner && p != owner {
		return Conversation{}, ErrParticipant
	}
	var ok bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1::uuid AND status='ACTIVE')`, p).Scan(&ok)
	if e != nil || !ok {
		return Conversation{}, ErrParticipant
	}
	a, b := u, p
	if a > b {
		a, b = b, a
	}
	var id string
	e = tx.QueryRow(ctx, `INSERT INTO conversations(job_id,participant_a,participant_b) VALUES($1,$2,$3) ON CONFLICT(job_id,participant_a,participant_b) DO UPDATE SET updated_at=conversations.updated_at RETURNING id::text`, j, a, b).Scan(&id)
	if e != nil {
		return Conversation{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return Conversation{}, e
	}
	return r.Get(ctx, u, id)
}
func (r *repository) Get(ctx context.Context, u, id string) (Conversation, error) {
	q := `SELECT ` + convCols + joins + `WHERE c.id=$1::uuid AND cm.user_id=$2::uuid`
	c, e := scanConv(r.p.QueryRow(ctx, q, id, u))
	if errors.Is(e, pgx.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, e
}
func page[T any](docs []T, total int64, i ListInput) Page[T] {
	p := Page[T]{Docs: docs, TotalDocs: total, Limit: i.Limit, Page: i.Page, PagingCounter: (i.Page-1)*i.Limit + 1}
	p.TotalPages = int((total + int64(i.Limit) - 1) / int64(i.Limit))
	p.HasPrevPage = i.Page > 1 && p.TotalPages > 0
	p.HasNextPage = i.Page < p.TotalPages
	if p.HasPrevPage {
		x := i.Page - 1
		p.PrevPage = &x
	}
	if p.HasNextPage {
		x := i.Page + 1
		p.NextPage = &x
	}
	return p
}
func (r *repository) List(ctx context.Context, u string, i ListInput) (Page[Conversation], error) {
	where := `cm.user_id=$1::uuid`
	args := []any{u}
	if i.Archived != nil {
		args = append(args, *i.Archived)
		where += fmt.Sprintf(` AND (cm.archived_at IS NOT NULL)=$%d`, len(args))
	}
	if i.Active != nil {
		args = append(args, *i.Active)
		where += fmt.Sprintf(` AND c.is_active=$%d`, len(args))
	}
	if i.Query != "" {
		args = append(args, "%"+i.Query+"%")
		where += fmt.Sprintf(` AND (concat_ws(' ',CASE WHEN c.participant_a=$1 THEN ub.first_name ELSE ua.first_name END,CASE WHEN c.participant_a=$1 THEN ub.last_name ELSE ua.last_name END) ILIKE $%d OR j.job_title ILIKE $%d)`, len(args), len(args))
	}
	var total int64
	if e := r.p.QueryRow(ctx, `SELECT count(*)`+joins+`WHERE `+where, args...).Scan(&total); e != nil {
		return Page[Conversation]{}, e
	}
	args = append(args, i.Limit, (i.Page-1)*i.Limit)
	rows, e := r.p.Query(ctx, `SELECT `+convCols+joins+`WHERE `+where+fmt.Sprintf(` ORDER BY c.last_message_at DESC,c.id DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if e != nil {
		return Page[Conversation]{}, e
	}
	defer rows.Close()
	ds := []Conversation{}
	for rows.Next() {
		c, e := scanConv(rows)
		if e != nil {
			return Page[Conversation]{}, e
		}
		ds = append(ds, c)
	}
	return page(ds, total, i), rows.Err()
}

const msgCols = `m.id::text,m.conversation_id::text,m.sender_id::text,m.receiver_id::text,m.message_type::text,coalesce(m.content,''),coalesce(m.file_url,''),coalesce(m.file_name,''),m.file_size,(m.read_at IS NOT NULL),m.read_at,m.delivered_at,(m.deleted_at IS NOT NULL),m.deleted_at,m.created_at,m.updated_at`

func scanMsg(s scan) (Message, error) {
	var m Message
	var sid, rid string
	e := s.Scan(&m.ID, &m.ConversationID, &sid, &rid, &m.MessageType, &m.Content, &m.FileURL, &m.FileName, &m.FileSize, &m.IsRead, &m.ReadAt, &m.DeliveredAt, &m.IsDeleted, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt)
	m.Sender = sid
	m.Receiver = rid
	m.IsDelivered = true
	return m, e
}
func (r *repository) Send(ctx context.Context, u, id string, in SendInput) (Message, error) {
	tx, e := r.p.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return Message{}, e
	}
	defer tx.Rollback(ctx)
	var recv string
	e = tx.QueryRow(ctx, `SELECT CASE WHEN participant_a=$2 THEN participant_b::text ELSE participant_a::text END FROM conversations WHERE id=$1 AND $2 IN (participant_a,participant_b) FOR UPDATE`, id, u).Scan(&recv)
	if errors.Is(e, pgx.ErrNoRows) || recv == "" {
		return Message{}, ErrNotFound
	}
	if e != nil {
		return Message{}, e
	}
	m, e := scanMsg(tx.QueryRow(ctx, `INSERT INTO messages(conversation_id,sender_id,receiver_id,message_type,content,file_url,file_name,file_size) VALUES($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8) RETURNING `+msgCols, id, u, recv, in.MessageType, in.Content, in.FileURL, in.FileName, in.FileSize))
	if e != nil {
		return m, e
	}
	_, e = tx.Exec(ctx, `UPDATE conversations SET last_message_id=$2,last_message_at=$3,updated_at=now() WHERE id=$1`, id, m.ID, m.CreatedAt)
	if e != nil {
		return m, e
	}
	_, e = tx.Exec(ctx, `UPDATE conversation_members SET unread_count=unread_count+1,archived_at=NULL WHERE conversation_id=$1 AND user_id=$2`, id, recv)
	if e != nil {
		return m, e
	}
	e = tx.Commit(ctx)
	return m, e
}
func (r *repository) Messages(ctx context.Context, u, id string, i ListInput) (Page[Message], error) {
	var exists bool
	if e := r.p.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id=$1 AND user_id=$2)`, id, u).Scan(&exists); e != nil {
		return Page[Message]{}, e
	}
	if !exists {
		return Page[Message]{}, ErrNotFound
	}
	var total int64
	r.p.QueryRow(ctx, `SELECT count(*) FROM messages WHERE conversation_id=$1 AND deleted_at IS NULL`, id).Scan(&total)
	rows, e := r.p.Query(ctx, `SELECT `+msgCols+` FROM messages m WHERE conversation_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, id, i.Limit, (i.Page-1)*i.Limit)
	if e != nil {
		return Page[Message]{}, e
	}
	defer rows.Close()
	ds := []Message{}
	for rows.Next() {
		m, e := scanMsg(rows)
		if e != nil {
			return Page[Message]{}, e
		}
		ds = append(ds, m)
	}
	return page(ds, total, i), rows.Err()
}
func (r *repository) MarkRead(ctx context.Context, u, id string) (int64, error) {
	tx, e := r.p.Begin(ctx)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback(ctx)
	var allowed bool
	e = tx.QueryRow(ctx, `SELECT true FROM conversations WHERE id=$1 AND $2 IN (participant_a,participant_b) FOR UPDATE`, id, u).Scan(&allowed)
	if errors.Is(e, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	if e != nil {
		return 0, e
	}
	tag, e := tx.Exec(ctx, `UPDATE messages SET read_at=now(),updated_at=now() WHERE conversation_id=$1 AND receiver_id=$2 AND read_at IS NULL AND deleted_at IS NULL`, id, u)
	if e != nil {
		return 0, e
	}
	tag2, e := tx.Exec(ctx, `UPDATE conversation_members SET unread_count=0,read_through_at=now() WHERE conversation_id=$1 AND user_id=$2`, id, u)
	if e != nil {
		return 0, e
	}
	if tag2.RowsAffected() == 0 {
		return 0, ErrNotFound
	}
	e = tx.Commit(ctx)
	return tag.RowsAffected(), e
}
func (r *repository) Delete(ctx context.Context, u, id string) (Message, error) {
	m, e := scanMsg(r.p.QueryRow(ctx, `UPDATE messages SET deleted_at=now(),updated_at=now() WHERE id=$1 AND sender_id=$2 AND deleted_at IS NULL RETURNING `+msgCols, id, u))
	if errors.Is(e, pgx.ErrNoRows) {
		return m, ErrMessageNotFound
	}
	return m, e
}
func (r *repository) SetArchived(ctx context.Context, u, id string, b bool) (Conversation, error) {
	tag, e := r.p.Exec(ctx, `UPDATE conversation_members SET archived_at=CASE WHEN $3 THEN now() ELSE NULL END WHERE conversation_id=$1 AND user_id=$2`, id, u, b)
	if e == nil && tag.RowsAffected() == 0 {
		return Conversation{}, ErrNotFound
	}
	if e != nil {
		return Conversation{}, e
	}
	return r.Get(ctx, u, id)
}
func (r *repository) MarkUrgent(ctx context.Context, u, id string) (Conversation, error) {
	tag, e := r.p.Exec(ctx, `UPDATE conversations c SET is_urgent=true,marked_urgent_by=$2,updated_at=now() WHERE id=$1 AND EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id=c.id AND user_id=$2)`, id, u)
	if e == nil && tag.RowsAffected() == 0 {
		return Conversation{}, ErrNotFound
	}
	if e != nil {
		return Conversation{}, e
	}
	return r.Get(ctx, u, id)
}
func (r *repository) Unread(ctx context.Context, u string) (int64, error) {
	var n int64
	e := r.p.QueryRow(ctx, `SELECT coalesce(sum(unread_count),0) FROM conversation_members WHERE user_id=$1`, u).Scan(&n)
	return n, e
}
