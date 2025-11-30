package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m04kA/SMC-SellerService/internal/domain"
	"github.com/m04kA/SMC-SellerService/pkg/dbmetrics"
	"github.com/m04kA/SMC-SellerService/pkg/psqlbuilder"

	"github.com/Masterminds/squirrel"
)

// Repository репозиторий для работы с услугами
type Repository struct {
	db DBExecutor
}

// NewRepository создает новый экземпляр репозитория услуг
func NewRepository(db DBExecutor) *Repository {
	return &Repository{db: db}
}

// Create создает новую услугу
func (r *Repository) Create(ctx context.Context, companyID int64, input domain.CreateServiceInput) (*domain.Service, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: Create - begin transaction: %v", ErrTransaction, err)
	}

	// Создаем услугу
	query, args, err := psqlbuilder.Insert("services").
		Columns("company_id", "name", "description", "average_duration").
		Values(companyID, input.Name, input.Description, input.AverageDuration).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: Create - build insert query: %v", ErrBuildQuery, err)
	}

	var serviceID int64
	var createdAt, updatedAt sql.NullTime
	err = tx.QueryRowContext(ctx, query, args...).Scan(&serviceID, &createdAt, &updatedAt)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: Create - insert service: %v", ErrExecQuery, err)
	}

	// Создаем связи с адресами
	for _, addressID := range input.AddressIDs {
		err = r.createServiceAddress(ctx, tx, serviceID, addressID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("Create - failed to create service address: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: Create - commit transaction: %v", ErrTransaction, err)
	}

	return &domain.Service{
		ID:              serviceID,
		CompanyID:       companyID,
		Name:            input.Name,
		Description:     input.Description,
		AverageDuration: input.AverageDuration,
		AddressIDs:      input.AddressIDs,
		CreatedAt:       createdAt.Time,
		UpdatedAt:       updatedAt.Time,
	}, nil
}

// GetByID получает услугу по ID
func (r *Repository) GetByID(ctx context.Context, companyID int64, serviceID int64) (*domain.Service, error) {
	query, args, err := psqlbuilder.Select("id", "company_id", "name", "description", "average_duration", "created_at", "updated_at").
		From("services").
		Where(squirrel.Eq{"id": serviceID, "company_id": companyID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%w: GetByID - build select query: %v", ErrBuildQuery, err)
	}

	var service domain.Service
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&service.ID,
		&service.CompanyID,
		&service.Name,
		&service.Description,
		&service.AverageDuration,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrServiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: GetByID - scan service: %v", ErrScanRow, err)
	}

	service.CreatedAt = createdAt.Time
	service.UpdatedAt = updatedAt.Time

	// Загружаем ID адресов
	addressIDs, err := r.getServiceAddressIDs(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("GetByID - failed to get address ids: %w", err)
	}
	service.AddressIDs = addressIDs

	return &service, nil
}

// ListByCompany получает список услуг компании
func (r *Repository) ListByCompany(ctx context.Context, companyID int64) ([]domain.Service, error) {
	query, args, err := psqlbuilder.Select("id", "company_id", "name", "description", "average_duration", "created_at", "updated_at").
		From("services").
		Where(squirrel.Eq{"company_id": companyID}).
		OrderBy("created_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	services := make([]domain.Service, 0)
	for rows.Next() {
		var service domain.Service
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&service.ID,
			&service.CompanyID,
			&service.Name,
			&service.Description,
			&service.AverageDuration,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		service.CreatedAt = createdAt.Time
		service.UpdatedAt = updatedAt.Time

		// Загружаем ID адресов для каждой услуги
		addressIDs, err := r.getServiceAddressIDs(ctx, service.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get address ids: %w", err)
		}
		service.AddressIDs = addressIDs

		services = append(services, service)
	}

	return services, nil
}

// Update обновляет услугу
func (r *Repository) Update(ctx context.Context, companyID int64, serviceID int64, input domain.UpdateServiceInput) (*domain.Service, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Обновляем основные поля услуги
	updateBuilder := psqlbuilder.Update("services").Where(squirrel.Eq{"id": serviceID, "company_id": companyID})

	if input.Name != nil {
		updateBuilder = updateBuilder.Set("name", *input.Name)
	}
	if input.Description != nil {
		updateBuilder = updateBuilder.Set("description", *input.Description)
	}
	if input.AverageDuration != nil {
		updateBuilder = updateBuilder.Set("average_duration", *input.AverageDuration)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("service not found")
	}

	// Обновляем адреса, если переданы
	if len(input.AddressIDs) > 0 {
		// Удаляем старые связи
		deleteQuery, deleteArgs, err := psqlbuilder.Delete("service_addresses").
			Where(squirrel.Eq{"service_id": serviceID}).
			ToSql()

		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to build delete service addresses query: %w", err)
		}

		_, err = tx.ExecContext(ctx, deleteQuery, deleteArgs...)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete old service addresses: %w", err)
		}

		// Создаем новые связи
		for _, addressID := range input.AddressIDs {
			err = r.createServiceAddress(ctx, tx, serviceID, addressID)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create service address: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Возвращаем обновленную услугу
	return r.GetByID(ctx, companyID, serviceID)
}

// Delete удаляет услугу
func (r *Repository) Delete(ctx context.Context, companyID int64, serviceID int64) error {
	query, args, err := psqlbuilder.Delete("services").
		Where(squirrel.Eq{"id": serviceID, "company_id": companyID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("%w: Delete - build delete query: %v", ErrBuildQuery, err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: Delete - execute delete: %v", ErrExecQuery, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: Delete - get rows affected: %v", ErrExecQuery, err)
	}

	if rowsAffected == 0 {
		return ErrServiceNotFound
	}

	return nil
}

// Helper methods

func (r *Repository) beginTx(ctx context.Context) (TxExecutor, error) {
	// Пытаемся привести к TxBeginner интерфейсу (dbmetrics.DB реализует этот интерфейс)
	if txBeginner, ok := r.db.(TxBeginner); ok {
		return txBeginner.BeginTx(ctx, nil)
	}

	// Fallback для обычного *sql.DB
	if db, ok := r.db.(*sql.DB); ok {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: beginTx: %v", ErrTransaction, err)
		}
		return &dbmetrics.SqlTxWrapper{Tx: tx}, nil
	}

	return nil, fmt.Errorf("%w: db type not supported", ErrTransaction)
}

func (r *Repository) createServiceAddress(ctx context.Context, tx TxExecutor, serviceID int64, addressID int64) error {
	query, args, err := psqlbuilder.Insert("service_addresses").
		Columns("service_id", "address_id").
		Values(serviceID, addressID).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert service address query: %w", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) getServiceAddressIDs(ctx context.Context, serviceID int64) ([]int64, error) {
	query, args, err := psqlbuilder.Select("address_id").
		From("service_addresses").
		Where(squirrel.Eq{"service_id": serviceID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addressIDs := make([]int64, 0)
	for rows.Next() {
		var addressID int64
		err := rows.Scan(&addressID)
		if err != nil {
			return nil, err
		}
		addressIDs = append(addressIDs, addressID)
	}

	return addressIDs, nil
}
