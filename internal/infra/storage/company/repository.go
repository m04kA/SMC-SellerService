package company

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m04kA/SMC-SellerService/internal/domain"
	"github.com/m04kA/SMC-SellerService/pkg/dbmetrics"
	"github.com/m04kA/SMC-SellerService/pkg/psqlbuilder"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

// Repository репозиторий для работы с компаниями
type Repository struct {
	db DBExecutor
}

// NewRepository создает новый экземпляр репозитория компаний
func NewRepository(db DBExecutor) *Repository {
	return &Repository{db: db}
}

// Create создает новую компанию
func (r *Repository) Create(ctx context.Context, input domain.CreateCompanyInput) (*domain.Company, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: Create - begin transaction: %v", ErrTransaction, err)
	}

	// Создаем компанию
	query, args, err := psqlbuilder.Insert("companies").
		Columns("name", "logo", "description", "tags", "manager_ids").
		Values(input.Name, input.Logo, input.Description, pq.Array(input.Tags), pq.Array(input.ManagerIDs)).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: Create - build insert query: %v", ErrBuildQuery, err)
	}

	var companyID int64
	var createdAt, updatedAt sql.NullTime
	err = tx.QueryRowContext(ctx, query, args...).Scan(&companyID, &createdAt, &updatedAt)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: Create - insert company: %v", ErrExecQuery, err)
	}

	// Создаем адреса
	addresses := make([]domain.Address, 0, len(input.Addresses))
	for _, addr := range input.Addresses {
		address, err := r.createAddress(ctx, tx, companyID, addr)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("Create - failed to create address: %w", err)
		}
		addresses = append(addresses, *address)
	}

	// Создаем рабочие часы
	err = r.createWorkingHours(ctx, tx, companyID, input.WorkingHours)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Create - failed to create working hours: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: Create - commit transaction: %v", ErrTransaction, err)
	}

	return &domain.Company{
		ID:           companyID,
		Name:         input.Name,
		Logo:         input.Logo,
		Description:  input.Description,
		Tags:         input.Tags,
		Addresses:    addresses,
		WorkingHours: input.WorkingHours,
		ManagerIDs:   input.ManagerIDs,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}, nil
}

// GetByID получает компанию по ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.Company, error) {
	query, args, err := psqlbuilder.Select("id", "name", "logo", "description", "tags", "manager_ids", "created_at", "updated_at").
		From("companies").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%w: GetByID - build select query: %v", ErrBuildQuery, err)
	}

	var company domain.Company
	var tags pq.StringArray
	var managerIDs pq.Int64Array
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&company.ID,
		&company.Name,
		&company.Logo,
		&company.Description,
		&tags,
		&managerIDs,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCompanyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: GetByID - scan company: %v", ErrScanRow, err)
	}

	company.CreatedAt = createdAt.Time
	company.UpdatedAt = updatedAt.Time
	company.Tags = tags
	company.ManagerIDs = managerIDs

	// Загружаем адреса
	addresses, err := r.getAddressesByCompanyID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID - failed to get addresses: %w", err)
	}
	company.Addresses = addresses

	// Загружаем рабочие часы
	workingHours, err := r.getWorkingHoursByCompanyID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID - failed to get working hours: %w", err)
	}
	company.WorkingHours = *workingHours

	return &company, nil
}


// List получает список компаний с фильтрацией
func (r *Repository) List(ctx context.Context, filter domain.CompanyFilter) ([]domain.Company, *domain.PaginationResult, error) {
	// Базовый запрос
	selectBuilder := psqlbuilder.Select("id", "name", "logo", "description", "tags", "manager_ids", "created_at", "updated_at").
		From("companies").
		OrderBy("created_at DESC")

	// Применяем фильтры
	if len(filter.Tags) > 0 {
		selectBuilder = selectBuilder.Where("tags && ?", pq.Array(filter.Tags))
	}

	if filter.City != nil {
		// Подзапрос для фильтрации по городу через таблицу адресов
		selectBuilder = selectBuilder.Where("id IN (SELECT company_id FROM addresses WHERE city = ?)", *filter.City)
	}

	// Применяем пагинацию только если Page и Limit заданы
	var pagination *domain.PaginationResult
	if filter.Page != nil && filter.Limit != nil {
		offset := (*filter.Page - 1) * *filter.Limit
		selectBuilder = selectBuilder.Limit(uint64(*filter.Limit)).Offset(uint64(offset))
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list companies: %w", err)
	}
	defer rows.Close()

	companies := make([]domain.Company, 0)
	for rows.Next() {
		var company domain.Company
		var tags pq.StringArray
		var managerIDs pq.Int64Array
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.Logo,
			&company.Description,
			&tags,
			&managerIDs,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan company: %w", err)
		}

		company.Tags = tags
		company.ManagerIDs = managerIDs
		company.CreatedAt = createdAt.Time
		company.UpdatedAt = updatedAt.Time

		// Загружаем адреса для каждой компании
		addresses, err := r.getAddressesByCompanyID(ctx, company.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get addresses: %w", err)
		}
		company.Addresses = addresses

		// Загружаем рабочие часы для каждой компании
		workingHours, err := r.getWorkingHoursByCompanyID(ctx, company.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get working hours: %w", err)
		}
		company.WorkingHours = *workingHours

		companies = append(companies, company)
	}

	// Получаем пагинацию только если она запрошена
	if filter.Page != nil && filter.Limit != nil {
		// Получаем общее количество
		countQuery, countArgs, err := psqlbuilder.Select("COUNT(*)").
			From("companies").
			ToSql()

		if err != nil {
			return nil, nil, fmt.Errorf("failed to build count query: %w", err)
		}

		var total int
		err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to count companies: %w", err)
		}

		pagination = &domain.PaginationResult{
			Page:  *filter.Page,
			Limit: *filter.Limit,
			Total: total,
		}
	}

	return companies, pagination, nil
}

// Update обновляет компанию
func (r *Repository) Update(ctx context.Context, id int64, input domain.UpdateCompanyInput) (*domain.Company, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: Update - begin transaction: %v", ErrTransaction, err)
	}

	// Обновляем основные поля компании
	updateBuilder := psqlbuilder.Update("companies").Where(squirrel.Eq{"id": id})

	if input.Name != nil {
		updateBuilder = updateBuilder.Set("name", *input.Name)
	}
	if input.Logo != nil {
		updateBuilder = updateBuilder.Set("logo", *input.Logo)
	}
	if input.Description != nil {
		updateBuilder = updateBuilder.Set("description", *input.Description)
	}
	if len(input.Tags) > 0 {
		updateBuilder = updateBuilder.Set("tags", pq.Array(input.Tags))
	}
	if len(input.ManagerIDs) > 0 {
		updateBuilder = updateBuilder.Set("manager_ids", pq.Array(input.ManagerIDs))
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: Update - build update query: %v", ErrBuildQuery, err)
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: Update - execute update: %v", ErrExecQuery, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: Update - get rows affected: %v", ErrExecQuery, err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return nil, ErrCompanyNotFound
	}

	// Обновляем адреса, если переданы
	if len(input.Addresses) > 0 {
		// Удаляем старые адреса
		deleteQuery, deleteArgs, err := psqlbuilder.Delete("addresses").
			Where(squirrel.Eq{"company_id": id}).
			ToSql()

		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("%w: Update - build delete addresses query: %v", ErrBuildQuery, err)
		}

		_, err = tx.ExecContext(ctx, deleteQuery, deleteArgs...)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("%w: Update - delete old addresses: %v", ErrExecQuery, err)
		}

		// Создаем новые адреса
		for _, addr := range input.Addresses {
			addressInput := domain.AddressInput{
				City:        addr.City,
				Street:      addr.Street,
				Building:    addr.Building,
				Coordinates: addr.Coordinates,
			}
			_, err := r.createAddress(ctx, tx, id, addressInput)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("Update - failed to create address: %w", err)
			}
		}
	}

	// Обновляем рабочие часы, если переданы
	if input.WorkingHours != nil {
		err = r.updateWorkingHours(ctx, tx, id, *input.WorkingHours)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("Update - failed to update working hours: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: Update - commit transaction: %v", ErrTransaction, err)
	}

	// Возвращаем обновленную компанию
	return r.GetByID(ctx, id)
}

// Delete удаляет компанию
func (r *Repository) Delete(ctx context.Context, id int64) error {
	query, args, err := psqlbuilder.Delete("companies").
		Where(squirrel.Eq{"id": id}).
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
		return ErrCompanyNotFound
	}

	return nil
}

// IsManager проверяет, является ли пользователь менеджером компании
func (r *Repository) IsManager(ctx context.Context, companyID int64, userID int64) (bool, error) {
	query, args, err := psqlbuilder.Select("manager_ids").
		From("companies").
		Where(squirrel.Eq{"id": companyID}).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("%w: IsManager - build query: %v", ErrBuildQuery, err)
	}

	var managerIDs pq.Int64Array
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&managerIDs)
	if err == sql.ErrNoRows {
		return false, ErrCompanyNotFound
	}
	if err != nil {
		return false, fmt.Errorf("%w: IsManager - scan manager ids: %v", ErrScanRow, err)
	}

	// Проверяем, есть ли userID в списке менеджеров
	for _, id := range managerIDs {
		if id == userID {
			return true, nil
		}
	}

	return false, nil
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

func (r *Repository) createAddress(ctx context.Context, tx TxExecutor, companyID int64, input domain.AddressInput) (*domain.Address, error) {
	query, args, err := psqlbuilder.Insert("addresses").
		Columns("company_id", "city", "street", "building", "latitude", "longitude").
		Values(companyID, input.City, input.Street, input.Building, input.Coordinates.Latitude, input.Coordinates.Longitude).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build insert address query: %w", err)
	}

	var address domain.Address
	var createdAt, updatedAt sql.NullTime

	err = tx.QueryRowContext(ctx, query, args...).Scan(&address.ID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	address.CompanyID = companyID
	address.City = input.City
	address.Street = input.Street
	address.Building = input.Building
	address.Coordinates = input.Coordinates
	address.CreatedAt = createdAt.Time
	address.UpdatedAt = updatedAt.Time

	return &address, nil
}

func (r *Repository) createWorkingHours(ctx context.Context, tx TxExecutor, companyID int64, wh domain.WorkingHours) error {
	query, args, err := psqlbuilder.Insert("working_hours").
		Columns(
			"company_id",
			"monday_is_open", "monday_open_time", "monday_close_time",
			"tuesday_is_open", "tuesday_open_time", "tuesday_close_time",
			"wednesday_is_open", "wednesday_open_time", "wednesday_close_time",
			"thursday_is_open", "thursday_open_time", "thursday_close_time",
			"friday_is_open", "friday_open_time", "friday_close_time",
			"saturday_is_open", "saturday_open_time", "saturday_close_time",
			"sunday_is_open", "sunday_open_time", "sunday_close_time",
		).
		Values(
			companyID,
			wh.Monday.IsOpen, wh.Monday.OpenTime, wh.Monday.CloseTime,
			wh.Tuesday.IsOpen, wh.Tuesday.OpenTime, wh.Tuesday.CloseTime,
			wh.Wednesday.IsOpen, wh.Wednesday.OpenTime, wh.Wednesday.CloseTime,
			wh.Thursday.IsOpen, wh.Thursday.OpenTime, wh.Thursday.CloseTime,
			wh.Friday.IsOpen, wh.Friday.OpenTime, wh.Friday.CloseTime,
			wh.Saturday.IsOpen, wh.Saturday.OpenTime, wh.Saturday.CloseTime,
			wh.Sunday.IsOpen, wh.Sunday.OpenTime, wh.Sunday.CloseTime,
		).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert working hours query: %w", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) updateWorkingHours(ctx context.Context, tx TxExecutor, companyID int64, wh domain.WorkingHours) error {
	query, args, err := psqlbuilder.Update("working_hours").
		Set("monday_is_open", wh.Monday.IsOpen).
		Set("monday_open_time", wh.Monday.OpenTime).
		Set("monday_close_time", wh.Monday.CloseTime).
		Set("tuesday_is_open", wh.Tuesday.IsOpen).
		Set("tuesday_open_time", wh.Tuesday.OpenTime).
		Set("tuesday_close_time", wh.Tuesday.CloseTime).
		Set("wednesday_is_open", wh.Wednesday.IsOpen).
		Set("wednesday_open_time", wh.Wednesday.OpenTime).
		Set("wednesday_close_time", wh.Wednesday.CloseTime).
		Set("thursday_is_open", wh.Thursday.IsOpen).
		Set("thursday_open_time", wh.Thursday.OpenTime).
		Set("thursday_close_time", wh.Thursday.CloseTime).
		Set("friday_is_open", wh.Friday.IsOpen).
		Set("friday_open_time", wh.Friday.OpenTime).
		Set("friday_close_time", wh.Friday.CloseTime).
		Set("saturday_is_open", wh.Saturday.IsOpen).
		Set("saturday_open_time", wh.Saturday.OpenTime).
		Set("saturday_close_time", wh.Saturday.CloseTime).
		Set("sunday_is_open", wh.Sunday.IsOpen).
		Set("sunday_open_time", wh.Sunday.OpenTime).
		Set("sunday_close_time", wh.Sunday.CloseTime).
		Where(squirrel.Eq{"company_id": companyID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update working hours query: %w", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) getAddressesByCompanyID(ctx context.Context, companyID int64) ([]domain.Address, error) {
	query, args, err := psqlbuilder.Select("id", "company_id", "city", "street", "building", "latitude", "longitude").
		From("addresses").
		Where(squirrel.Eq{"company_id": companyID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select addresses query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make([]domain.Address, 0)
	for rows.Next() {
		var addr domain.Address

		err := rows.Scan(
			&addr.ID,
			&addr.CompanyID,
			&addr.City,
			&addr.Street,
			&addr.Building,
			&addr.Coordinates.Latitude,
			&addr.Coordinates.Longitude,
		)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}

func (r *Repository) getWorkingHoursByCompanyID(ctx context.Context, companyID int64) (*domain.WorkingHours, error) {
	query, args, err := psqlbuilder.Select(
		"monday_is_open", "monday_open_time", "monday_close_time",
		"tuesday_is_open", "tuesday_open_time", "tuesday_close_time",
		"wednesday_is_open", "wednesday_open_time", "wednesday_close_time",
		"thursday_is_open", "thursday_open_time", "thursday_close_time",
		"friday_is_open", "friday_open_time", "friday_close_time",
		"saturday_is_open", "saturday_open_time", "saturday_close_time",
		"sunday_is_open", "sunday_open_time", "sunday_close_time",
	).
		From("working_hours").
		Where(squirrel.Eq{"company_id": companyID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select working hours query: %w", err)
	}

	var wh domain.WorkingHours
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&wh.Monday.IsOpen, &wh.Monday.OpenTime, &wh.Monday.CloseTime,
		&wh.Tuesday.IsOpen, &wh.Tuesday.OpenTime, &wh.Tuesday.CloseTime,
		&wh.Wednesday.IsOpen, &wh.Wednesday.OpenTime, &wh.Wednesday.CloseTime,
		&wh.Thursday.IsOpen, &wh.Thursday.OpenTime, &wh.Thursday.CloseTime,
		&wh.Friday.IsOpen, &wh.Friday.OpenTime, &wh.Friday.CloseTime,
		&wh.Saturday.IsOpen, &wh.Saturday.OpenTime, &wh.Saturday.CloseTime,
		&wh.Sunday.IsOpen, &wh.Sunday.OpenTime, &wh.Sunday.CloseTime,
	)

	if err != nil {
		return nil, err
	}

	return &wh, nil
}
