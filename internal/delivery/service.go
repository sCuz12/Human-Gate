package delivery

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"humangate/db/generated"
	"humangate/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultBatchSize    = 10
	maxDeliveryAttempts = 6
)

var retryDelays = []time.Duration{
	5 * time.Second,
	30 * time.Second,
	2 * time.Minute,
	10 * time.Minute,
	1 * time.Hour,
}

type Service struct {
	db         *pgxpool.Pool
	httpClient *http.Client
	logger     *slog.Logger
	signingKey []byte
	clock      func() time.Time
}

type Config struct {
	SigningKey string
}

func NewService(db *pgxpool.Pool, httpClient *http.Client, logger *slog.Logger, cfg Config) *Service {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		db:         db,
		httpClient: httpClient,
		logger:     logger,
		signingKey: []byte(cfg.SigningKey),
		clock:      time.Now,
	}
}

func (s *Service) ProcessDue(ctx context.Context) error {
	now := s.clock().UTC()
	queries := generated.New(s.db)

	deliveries, err := queries.ListDueDecisionDeliveries(ctx, generated.ListDueDecisionDeliveriesParams{
		NowAt:      pgxutil.Timestamptz(now),
		LimitCount: defaultBatchSize,
	})
	if err != nil {
		return fmt.Errorf("list due decision deliveries: %w", err)
	}

	for _, delivery := range deliveries {
		fmt.Println(delivery.Destination)
		if err := s.processOne(ctx, queries, delivery); err != nil {
			s.logger.ErrorContext(
				ctx,
				"process decision delivery failed",
				"error", err,
				"delivery_id", pgxutil.UUIDString(delivery.DeliveryID),
				"workspace_id", pgxutil.UUIDString(delivery.WorkspaceID),
				"decision_id", pgxutil.UUIDString(delivery.DecisionID),
			)
		}
	}

	return nil
}

func (s *Service) processOne(ctx context.Context, queries *generated.Queries, delivery generated.ListDueDecisionDeliveriesRow) error {
	now := s.clock().UTC()
	attemptCount := delivery.AttemptCount + 1

	if _, err := queries.UpdateDecisionDeliveryAttempt(ctx, generated.UpdateDecisionDeliveryAttemptParams{
		WorkspaceID:      delivery.WorkspaceID,
		ID:               delivery.DeliveryID,
		Status:           generated.DeliveryStatusAttempting,
		AttemptCount:     attemptCount,
		NextAttemptAt:    pgtype.Timestamptz{},
		LastAttemptAt:    pgxutil.Timestamptz(now),
		LastResponseCode: pgtype.Int4{},
		LastError:        pgtype.Text{},
		DeliveredAt:      pgtype.Timestamptz{},
		AcknowledgedAt:   pgtype.Timestamptz{},
		UpdatedAt:        pgxutil.Timestamptz(now),
	}); err != nil {
		return fmt.Errorf("mark delivery attempting: %w", err)
	}

	statusCode, deliverErr := s.deliver(ctx, delivery)
	if deliverErr == nil && statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		_, err := queries.UpdateDecisionDeliveryAttempt(ctx, generated.UpdateDecisionDeliveryAttemptParams{
			WorkspaceID:      delivery.WorkspaceID,
			ID:               delivery.DeliveryID,
			Status:           generated.DeliveryStatusDelivered,
			AttemptCount:     attemptCount,
			NextAttemptAt:    pgtype.Timestamptz{},
			LastAttemptAt:    pgxutil.Timestamptz(now),
			LastResponseCode: int4(statusCode),
			LastError:        pgtype.Text{},
			DeliveredAt:      pgxutil.Timestamptz(now),
			AcknowledgedAt:   pgtype.Timestamptz{},
			UpdatedAt:        pgxutil.Timestamptz(now),
		})
		if err != nil {
			return fmt.Errorf("mark delivery delivered: %w", err)
		}

		s.logger.InfoContext(
			ctx,
			"decision delivered",
			"delivery_id", pgxutil.UUIDString(delivery.DeliveryID),
			"workspace_id", pgxutil.UUIDString(delivery.WorkspaceID),
			"decision_id", pgxutil.UUIDString(delivery.DecisionID),
			"status_code", statusCode,
		)
		return nil
	}

	errMessage := "non-2xx response"
	if deliverErr != nil {
		errMessage = deliverErr.Error()
	}

	nextStatus := generated.DeliveryStatusRetrying
	nextAttemptAt := pgxutil.Timestamptz(now.Add(nextRetryDelay(attemptCount)))
	if attemptCount >= maxDeliveryAttempts {
		nextStatus = generated.DeliveryStatusPermanentlyFailed
		nextAttemptAt = pgtype.Timestamptz{}
	}

	if _, err := queries.UpdateDecisionDeliveryAttempt(ctx, generated.UpdateDecisionDeliveryAttemptParams{
		WorkspaceID:      delivery.WorkspaceID,
		ID:               delivery.DeliveryID,
		Status:           nextStatus,
		AttemptCount:     attemptCount,
		NextAttemptAt:    nextAttemptAt,
		LastAttemptAt:    pgxutil.Timestamptz(now),
		LastResponseCode: int4(statusCode),
		LastError:        text(errMessage),
		DeliveredAt:      pgtype.Timestamptz{},
		AcknowledgedAt:   pgtype.Timestamptz{},
		UpdatedAt:        pgxutil.Timestamptz(now),
	}); err != nil {
		return fmt.Errorf("mark delivery failed: %w", err)
	}

	return fmt.Errorf("deliver decision: %s", errMessage)
}

func (s *Service) deliver(ctx context.Context, delivery generated.ListDueDecisionDeliveriesRow) (int, error) {
	if delivery.Strategy != generated.ContinuationStrategyWebhook {
		return 0, fmt.Errorf("unsupported continuation strategy %q", delivery.Strategy)
	}
	if !delivery.Destination.Valid || strings.TrimSpace(delivery.Destination.String) == "" {
		return 0, fmt.Errorf("missing continuation destination")
	}

	body, err := s.signedPayload(delivery)
	if err != nil {
		return 0, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, delivery.Destination.String, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build delivery request: %w", err)
	}

	signature := sign(body, s.signingKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "HumanGate-Worker/0.1")
	request.Header.Set("X-HumanGate-Signature", "sha256="+signature)
	request.Header.Set("X-HumanGate-Decision-ID", pgxutil.UUIDString(delivery.DecisionID))
	request.Header.Set("X-HumanGate-Delivery-ID", pgxutil.UUIDString(delivery.DeliveryID))

	response, err := s.httpClient.Do(request)
	if err != nil {
		return 0, fmt.Errorf("send delivery request: %w", err)
	}
	defer response.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response.StatusCode, fmt.Errorf("callback returned status %d", response.StatusCode)
	}

	return response.StatusCode, nil
}

func (s *Service) signedPayload(delivery generated.ListDueDecisionDeliveriesRow) ([]byte, error) {
	payload := map[string]any{
		"delivery_id":           pgxutil.UUIDString(delivery.DeliveryID),
		"workspace_id":          pgxutil.UUIDString(delivery.WorkspaceID),
		"approval_request_id":   pgxutil.UUIDString(delivery.ApprovalRequestID),
		"decision_id":           pgxutil.UUIDString(delivery.DecisionID),
		"decision":              delivery.Decision,
		"action_type":           delivery.ActionType,
		"original_action":       rawJSON(delivery.OriginalAction),
		"original_action_hash":  delivery.OriginalActionHash,
		"approved_action":       rawJSON(delivery.ApprovedAction),
		"approved_action_hash":  nullableText(delivery.ApprovedActionHash),
		"source_platform":       delivery.Platform,
		"continuation_strategy": delivery.Strategy,
		"issued_at":             delivery.IssuedAt.Time.UTC().Format(time.RFC3339Nano),
		"signature_algorithm":   "HMAC-SHA256",
		"signature_header":      "X-HumanGate-Signature",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal delivery payload: %w", err)
	}

	return body, nil
}

func sign(body []byte, key []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func nextRetryDelay(attemptCount int32) time.Duration {
	index := int(attemptCount) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(retryDelays) {
		return retryDelays[len(retryDelays)-1]
	}
	return retryDelays[index]
}

func rawJSON(value []byte) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage("null")
	}
	return json.RawMessage(value)
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func int4(value int) pgtype.Int4 {
	if value == 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(value), Valid: true}
}

func text(value string) pgtype.Text {
	if strings.TrimSpace(value) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
