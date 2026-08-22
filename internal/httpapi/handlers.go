package httpapi

import (
	"encoding/json"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type Protected struct{ Services *service.Services }

func (p Protected) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/fields", p.createField)
	mux.HandleFunc("POST /api/v1/wells", p.createWell)
	mux.HandleFunc("GET /api/v1/summary", p.summary)
}
func (p Protected) actor(r *http.Request) string {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token
}
func (p Protected) createField(w http.ResponseWriter, r *http.Request) {
	actor := p.actor(r)
	var input domain.Field
	if e := json.NewDecoder(r.Body).Decode(&input); e != nil {
		write(w, 400, errorBody(domain.ErrInvalid, r))
		return
	}
	field, e := p.Services.Operations.CreateField(r.Context(), actor, input)
	if e != nil {
		write(w, 409, errorBody(e, r))
		return
	}
	write(w, 201, field)
}
func (p Protected) createWell(w http.ResponseWriter, r *http.Request) {
	actor := p.actor(r)
	var input domain.Well
	if e := json.NewDecoder(r.Body).Decode(&input); e != nil {
		write(w, 400, errorBody(domain.ErrInvalid, r))
		return
	}
	well, e := p.Services.Operations.CreateWell(r.Context(), actor, input)
	if e != nil {
		write(w, 409, errorBody(e, r))
		return
	}
	write(w, 201, well)
}
func (p Protected) summary(w http.ResponseWriter, r *http.Request) {
	summary, e := p.Services.Operations.Store.Summary(r.Context())
	if e != nil {
		write(w, 500, errorBody(e, r))
		return
	}
	if value := r.URL.Query().Get("limit"); value != "" {
		_, _ = strconv.Atoi(value)
	}
	write(w, 200, summary)
}
