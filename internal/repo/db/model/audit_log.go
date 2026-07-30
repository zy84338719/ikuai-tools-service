package model

import "time"

// AuditLog records a mutating operation (POST/PUT/PATCH/DELETE) against the
// service for traceability. Read-only requests are not audited.
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Actor     string    `gorm:"size:100;index" json:"actor"`           // api-key id / username / "anonymous"
	Method    string    `gorm:"size:10;index" json:"method"`           // HTTP method
	Path      string    `gorm:"size:255;index" json:"path"`            // request path
	RouterID  string    `gorm:"size:50;index" json:"router_id"`        // target router ("" when N/A)
	Status    int       `json:"status"`                                // response status code
	ReqID     string    `gorm:"size:64;index" json:"req_id"`           // X-Request-ID
	IP        string    `gorm:"size:64" json:"ip"`                     // client IP
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
