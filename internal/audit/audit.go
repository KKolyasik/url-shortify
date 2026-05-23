package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/KKolyasik/url-shortify/internal/config"
	"github.com/KKolyasik/url-shortify/internal/model"
	"go.uber.org/zap"
)

type Audit struct {
	auditCfg   *config.AuditConfig
	httpClient *http.Client
	logger     *zap.Logger
}

func NewAudit(auditCfg *config.AuditConfig, logger *zap.Logger) *Audit {
	return &Audit{
		auditCfg:   auditCfg,
		httpClient: &http.Client{Timeout: time.Second * 10},
		logger:     logger,
	}
}

func (a *Audit) Notify(ctx context.Context, event model.AuditEvent) {
	if a.auditCfg.AuditURL != "" {
		if err := a.sendRequest(ctx, event); err != nil {
			a.logger.Warn("failed to send audit request", zap.Error(err))
		}
	}

	if a.auditCfg.AuditFile != "" {
		if err := a.writeToFile(event); err != nil {
			a.logger.Warn("failed to write audit to file", zap.Error(err))
		}
	}
}

func (a *Audit) sendRequest(ctx context.Context, event model.AuditEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshall audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.auditCfg.AuditURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create audit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do audit request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (a *Audit) writeToFile(event model.AuditEvent) error {
	file, err := os.OpenFile(a.auditCfg.AuditFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("filed to open file: %w", err)
	}
	defer file.Close()

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshall audit event: %w", err)
	}

	if _, err := file.Write(append(body, '\n')); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	return nil
}
