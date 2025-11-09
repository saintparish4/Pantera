package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/domain"
)

// CalculationLogRepo implements domain.CalculationLogRepository
type CalculationLogRepo struct {
	db *sql.DB
}

// NewCalculationLogRepository creates a new calculation log repository
func NewCalculationLogRepository(db *sql.DB) domain.CalculationLogRepository {
	return &CalculationLogRepo{db: db}
}

// Create creates a new calculation log entry
func (r *CalculationLogRepo) Create(ctx context.Context, log *domain.CalculationLog) error {
	query := `
		INSERT INTO calculation_logs (
			id, user_id, api_key_id, rule_id, strategy_type, 
			input_data, output_data, execution_time_ms, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Generate ID if not provided
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	// Set timestamp
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// Convert data to JSONB
	inputData := FromMap(log.InputData)
	outputData := FromMap(log.OutputData)

	_, err := r.db.ExecContext(
		ctx,
		query,
		log.ID,
		log.UserID,
		log.APIKeyID,
		log.RuleID,
		log.StrategyType,
		inputData,
		outputData,
		log.ExecutionTimeMs,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create calculation log: %w", err)
	}

	return nil
}

// GetByID retrieves a calculation log by ID
func (r *CalculationLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CalculationLog, error) {
	query := `
		SELECT id, user_id, api_key_id, rule_id, strategy_type, 
		       input_data, output_data, execution_time_ms, created_at
		FROM calculation_logs
		WHERE id = $1
	`

	log := &domain.CalculationLog{}
	var inputData, outputData JSONB
	var apiKeyID, ruleID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.UserID,
		&apiKeyID,
		&ruleID,
		&log.StrategyType,
		&inputData,
		&outputData,
		&log.ExecutionTimeMs,
		&log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("calculation log not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get calculation log: %w", err)
	}

	// Convert JSONB to maps
	log.InputData = inputData.ToMap()
	log.OutputData = outputData.ToMap()

	// Handle nullable UUIDs
	if apiKeyID.Valid {
		keyID, err := uuid.Parse(apiKeyID.String)
		if err == nil {
			log.APIKeyID = &keyID
		}
	}

	if ruleID.Valid {
		rID, err := uuid.Parse(ruleID.String)
		if err == nil {
			log.RuleID = &rID
		}
	}

	return log, nil
}

// GetByUserID retrieves calculation logs for a user with pagination
func (r *CalculationLogRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.CalculationLog, error) {
	query := `
		SELECT id, user_id, api_key_id, rule_id, strategy_type, 
		       input_data, output_data, execution_time_ms, created_at
		FROM calculation_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query calculation logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.CalculationLog

	for rows.Next() {
		log := &domain.CalculationLog{}
		var inputData, outputData JSONB
		var apiKeyID, ruleID sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&apiKeyID,
			&ruleID,
			&log.StrategyType,
			&inputData,
			&outputData,
			&log.ExecutionTimeMs,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calculation log: %w", err)
		}

		log.InputData = inputData.ToMap()
		log.OutputData = outputData.ToMap()

		if apiKeyID.Valid {
			keyID, err := uuid.Parse(apiKeyID.String)
			if err == nil {
				log.APIKeyID = &keyID
			}
		}

		if ruleID.Valid {
			rID, err := uuid.Parse(ruleID.String)
			if err == nil {
				log.RuleID = &rID
			}
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calculation logs: %w", err)
	}

	return logs, nil
}

// List retrieves calculation logs with filters
func (r *CalculationLogRepo) List(ctx context.Context, filter domain.CalculationLogFilter) ([]*domain.CalculationLog, error) {
	query := `
		SELECT id, user_id, api_key_id, rule_id, strategy_type, 
		       input_data, output_data, execution_time_ms, created_at
		FROM calculation_logs
		WHERE user_id = $1
	`

	args := []interface{}{filter.UserID}
	argCount := 1

	// Add API key filter
	if filter.APIKeyID != nil {
		argCount++
		query += fmt.Sprintf(" AND api_key_id = $%d", argCount)
		args = append(args, *filter.APIKeyID)
	}

	// Add rule ID filter
	if filter.RuleID != nil {
		argCount++
		query += fmt.Sprintf(" AND rule_id = $%d", argCount)
		args = append(args, *filter.RuleID)
	}

	// Add strategy type filter
	if filter.StrategyType != "" {
		argCount++
		query += fmt.Sprintf(" AND strategy_type = $%d", argCount)
		args = append(args, filter.StrategyType)
	}

	// Add time range filters
	if !filter.From.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filter.From)
	}

	if !filter.To.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filter.To)
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list calculation logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.CalculationLog

	for rows.Next() {
		log := &domain.CalculationLog{}
		var inputData, outputData JSONB
		var apiKeyID, ruleID sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&apiKeyID,
			&ruleID,
			&log.StrategyType,
			&inputData,
			&outputData,
			&log.ExecutionTimeMs,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calculation log: %w", err)
		}

		log.InputData = inputData.ToMap()
		log.OutputData = outputData.ToMap()

		if apiKeyID.Valid {
			keyID, err := uuid.Parse(apiKeyID.String)
			if err == nil {
				log.APIKeyID = &keyID
			}
		}

		if ruleID.Valid {
			rID, err := uuid.Parse(ruleID.String)
			if err == nil {
				log.RuleID = &rID
			}
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calculation logs: %w", err)
	}

	return logs, nil
}

// GetStats retrieves calculation statistics for a user
func (r *CalculationLogRepo) GetStats(ctx context.Context, userID uuid.UUID, from, to time.Time) (*domain.CalculationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_calculations,
			AVG(execution_time_ms) as avg_execution_time,
			MIN((output_data->>'final_price')::numeric) as min_price,
			MAX((output_data->>'final_price')::numeric) as max_price,
			AVG((output_data->>'final_price')::numeric) as avg_price
		FROM calculation_logs
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
	`

	stats := &domain.CalculationStats{
		ByStrategy: make(map[string]int),
	}

	var minPrice, maxPrice, avgPrice sql.NullFloat64
	var avgExecutionTime sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, userID, from, to).Scan(
		&stats.TotalCalculations,
		&avgExecutionTime,
		&minPrice,
		&maxPrice,
		&avgPrice,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get calculation stats: %w", err)
	}

	// Handle nullable values
	if avgExecutionTime.Valid {
		stats.AvgExecutionTime = avgExecutionTime.Float64
	}
	if minPrice.Valid {
		stats.MinPrice = minPrice.Float64
	}
	if maxPrice.Valid {
		stats.MaxPrice = maxPrice.Float64
	}
	if avgPrice.Valid {
		stats.AvgPrice = avgPrice.Float64
	}

	// Get strategy breakdown
	strategyQuery := `
		SELECT strategy_type, COUNT(*) as count
		FROM calculation_logs
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY strategy_type
	`

	rows, err := r.db.QueryContext(ctx, strategyQuery, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var strategy string
		var count int

		if err := rows.Scan(&strategy, &count); err != nil {
			return nil, fmt.Errorf("failed to scan strategy breakdown: %w", err)
		}

		stats.ByStrategy[strategy] = count
	}

	// Set period
	stats.Period = fmt.Sprintf("%s to %s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	return stats, nil
}
