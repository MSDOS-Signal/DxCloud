// Package dto 定义请求/响应结构体。
package dto

// ---------- 认证 ----------

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserInfo struct {
	ID          uint64   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Nickname    string   `json:"nickname"`
	AvatarURL   string   `json:"avatar_url"`
	Status      int8     `json:"status"`
	Roles       []string `json:"roles"`
	RoleNames   []string `json:"role_names"`
	Permissions []string `json:"permissions"`
}

// ---------- IAM 管理 ----------

type CreateUserReq struct {
	Username  string   `json:"username" binding:"required"`
	Email     string   `json:"email" binding:"required,email"`
	Password  string   `json:"password" binding:"required"`
	Nickname  string   `json:"nickname"`
	RoleCodes []string `json:"role_codes"`
}

type UpdateUserReq struct {
	Nickname string `json:"nickname"`
	Status   *int8  `json:"status"`
}

type GrantRolesReq struct {
	RoleCodes []string `json:"role_codes" binding:"required"`
}

type CreateRoleReq struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
}

type UpdateRoleReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
}

type GrantPermsReq struct {
	PermCodes []string `json:"perm_codes" binding:"required"`
}

type PageResult struct {
	Total int64 `json:"total"`
	Items any   `json:"items"`
}
