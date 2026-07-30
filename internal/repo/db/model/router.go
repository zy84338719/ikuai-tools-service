package model

import "time"

// Router represents a managed iKuai router instance. The Manager registry
// builds one *ikuaiapi.Client per row. Tokens are stored at rest; protect the
// database accordingly in production.
type Router struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:50;not null" json:"name"` // human label, also the router_id used in paths
	BaseURL   string    `gorm:"size:255;not null" json:"base_url"`
	Token     string    `gorm:"size:128;not null" json:"-"`               // hidden from JSON responses
	Insecure  bool      `gorm:"default:true" json:"insecure"`
	Timeout   int       `gorm:"default:30" json:"timeout"`               // seconds
	Status    int8      `gorm:"default:1" json:"status"`                 // 1=enabled, 0=disabled
	Comment   string    `gorm:"size:255" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Router) TableName() string {
	return "routers"
}

// RouterResp is the token-redacted view returned by the API.
type RouterResp struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	Insecure  bool      `json:"insecure"`
	Timeout   int       `json:"timeout"`
	Status    int8      `json:"status"`
	Comment   string    `json:"comment"`
	HasToken  bool      `json:"has_token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *Router) ToResp() *RouterResp {
	return &RouterResp{
		ID:        r.ID,
		Name:      r.Name,
		BaseURL:   r.BaseURL,
		Insecure:  r.Insecure,
		Timeout:   r.Timeout,
		Status:    r.Status,
		Comment:   r.Comment,
		HasToken:  r.Token != "",
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
