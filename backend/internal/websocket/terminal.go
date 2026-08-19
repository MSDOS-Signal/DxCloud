// Package websocket：Web Terminal（浏览器 xterm.js ↔ WSS ↔ docker exec TTY）。
// 鉴权链：一次性令牌（Redis 60s，绑定 user+instance，用后即焚）→ RBAC ecs:console → 属主校验。
// 安全：exec 会话绑定单容器；容器侧默认非特权 + no-new-privileges（见 docker provider 基线）。
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	idleTimeout = 15 * time.Minute // 空闲超时（无输入自动断开）
	writeWait   = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 鉴权不依赖 Origin（一次性令牌），生产由 Traefik 同域收敛；开发期放开
	CheckOrigin: func(r *http.Request) bool { return true },
}

type controlMsg struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type TerminalHandler struct {
	svc    *service.EcsService
	iamSvc *iam.Service
	log    *zap.Logger
}

func NewTerminalHandler(svc *service.EcsService, iamSvc *iam.Service, log *zap.Logger) *TerminalHandler {
	return &TerminalHandler{svc: svc, iamSvc: iamSvc, log: log}
}

func fail(c *gin.Context, status, code int, msg string) {
	c.JSON(status, gin.H{"code": code, "message": msg})
}

// Handle GET /ws/v1/ecs/:id/terminal?token=xxx
func (h *TerminalHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// 1) 一次性令牌（用后即焚）
	userID, instanceID, err := h.svc.ValidateConsoleToken(ctx, c.Query("token"))
	if err != nil {
		fail(c, http.StatusUnauthorized, 40100, "invalid or expired token")
		return
	}

	// 2) 二次鉴权：权限 ecs:console（令牌只是门票，这里才是闸门）
	roles, err := h.iamSvc.GetUserRoleCodes(ctx, userID)
	if err != nil {
		fail(c, http.StatusUnauthorized, 40100, "unauthorized")
		return
	}
	ac := service.AccessCtx{UserID: userID, Roles: roles}
	hasPerm, _ := h.iamSvc.HasPerm(ctx, userID, "ecs:console")
	if !hasPerm {
		h.iamSvc.Audit(ctx, &userID, "console.deny", "ecs", fmtUint(instanceID), c.ClientIP(), "", 2, map[string]any{"reason": "no ecs:console"})
		fail(c, http.StatusForbidden, 40001, "没有操作权限")
		return
	}

	// 3) 属主/状态校验（与 REST 同一套 accessCheck）
	inst, err := h.svc.ConsoleAccessCheck(ctx, ac, instanceID)
	if err != nil {
		if err == service.ErrNotFound {
			fail(c, http.StatusNotFound, 40400, "instance not found")
		} else if err == service.ErrForbidden {
			h.iamSvc.Audit(ctx, &userID, "console.deny", "ecs", fmtUint(instanceID), c.ClientIP(), "", 2, map[string]any{"reason": "not owner"})
			fail(c, http.StatusForbidden, 40001, "没有操作权限")
		} else {
			fail(c, http.StatusBadRequest, 40000, err.Error())
		}
		return
	}

	// 4) WebSocket 升级
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	h.iamSvc.Audit(ctx, &userID, "console.open", "ecs", inst.InstanceNo, c.ClientIP(), "", 1, nil)
	start := time.Now()
	defer func() {
		h.iamSvc.Audit(ctx, &userID, "console.close", "ecs", inst.InstanceNo, c.ClientIP(), "", 1,
			map[string]any{"duration_sec": time.Since(start).Seconds()})
	}()

	h.bridge(ctx, ws, inst.ContainerID)
}

func fmtUint(v uint64) string {
	return fmt.Sprint(v)
}

// bridge ws ↔ exec PTY 双向桥接：二进制帧=终端 I/O，文本帧=JSON 控制（resize）。
func (h *TerminalHandler) bridge(ctx context.Context, ws *websocket.Conn, containerID string) {
	execID, err := h.svc.ExecCreate(ctx, containerID)
	if err != nil {
		h.log.Warn("exec create failed", zap.Error(err))
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "exec create failed"), time.Now().Add(writeWait))
		return
	}
	sess, err := h.svc.ExecAttach(ctx, containerID, execID)
	if err != nil {
		h.log.Warn("exec attach failed", zap.Error(err))
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "exec attach failed"), time.Now().Add(writeWait))
		return
	}
	defer sess.Close()

	// exec → ws
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := sess.Reader.Read(buf)
			if n > 0 {
				h.log.Debug("exec output", zap.Int("bytes", n))
				_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
				if werr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					h.log.Debug("ws write failed, output loop exiting", zap.Error(werr))
					return
				}
			}
			if err != nil {
				h.log.Debug("exec read ended", zap.Error(err))
				// 会话结束：通知对端关闭
				_ = ws.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session ended"), time.Now().Add(writeWait))
				return
			}
		}
	}()

	// ws → exec
	_ = ws.SetReadDeadline(time.Now().Add(idleTimeout))
	for {
		mt, data, err := ws.ReadMessage()
		if err != nil {
			return
		}
		switch mt {
		case websocket.BinaryMessage:
			if len(data) > 0 {
				if _, err := sess.Conn.Write(data); err != nil {
					return
				}
			}
		case websocket.TextMessage:
			var msg controlMsg
			if json.Unmarshal(data, &msg) == nil && msg.Type == "resize" && msg.Cols > 0 && msg.Rows > 0 {
				if err := h.svc.ExecResize(ctx, execID, msg.Cols, msg.Rows); err != nil {
					h.log.Debug("exec resize failed", zap.Error(err))
				}
			}
		case websocket.PingMessage:
			_ = ws.WriteMessage(websocket.PongMessage, nil)
		case websocket.CloseMessage:
			return
		}
		_ = ws.SetReadDeadline(time.Now().Add(idleTimeout))
	}
}
