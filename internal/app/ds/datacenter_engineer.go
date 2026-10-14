package ds

import "time"

type DatacenterEngineer struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Login       string `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}

type DatacenterEngineerSignUpRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type DatacenterEngineerSignInRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type DatacenterEngineerSignInResponse struct {
	ExpiresIn   time.Duration `json:"expires_in" swaggertype:"integer"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
}
