package audit

import (
	"context"
	"encoding/json"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/identity"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"github.com/zhanglei10281852-gif/newgo/internal/requestmeta"
	"time"
)

type Recorder struct {
	Store repository.Store
	Now   func() time.Time
}

func (r Recorder) Record(ctx context.Context, actor, action, kind, id, outcome string, meta map[string]string) error {
	b, _ := json.Marshal(meta)
	return r.Store.InsertAudit(ctx, domain.AuditEvent{ID: identity.New("audit"), ActorID: actor, Action: action, EntityType: kind, EntityID: id, Outcome: outcome, RequestID: requestmeta.ID(ctx), Metadata: string(b), CreatedAt: r.Now()})
}
