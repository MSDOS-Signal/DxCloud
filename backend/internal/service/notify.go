// Package service：站内通知助手（各模块操作结果统一落通知中心，带跳转链接）。
package service

import (
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
)

// notify 发送站内通知（失败不影响主流程，仅记录）。
func notify(repo *repository.Repos, userID uint64, typ, title, content, link string) {
	if userID == 0 {
		return
	}
	_ = repo.NotifyCreate(&model.Notification{
		UserID: userID, Type: typ, Title: title, Content: content, Link: link,
	})
}
