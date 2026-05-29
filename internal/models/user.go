package models

import "time"

type User struct {
    ID          string    `json:"id"`
    FirebaseUID string    `json:"-"` // "-" means never include in JSON output
    Email       string    `json:"email"`
    DisplayName string    `json:"display_name"`
    AvatarURL   string    `json:"avatar_url,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}
