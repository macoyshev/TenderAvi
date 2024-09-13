package tender

import (
	"database/sql"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"avi/internal/database"
	"avi/internal/model"

	"github.com/google/uuid"
)

type TenderRepo struct {
	db *sql.DB
}

func (repo *TenderRepo) CreateTender(
	name string,
	description string,
	serviceType model.TenderServiceType,
	organizarionId uuid.UUID,
	userId uuid.UUID,
) (tender *model.Tender, err error) {
	var id uuid.UUID
	var version int32
	var createdAt time.Time
	var status model.TenderStatus

	createQuery := `
		INSERT INTO tender 
		(name, description, service_type, 
		organization_id, user_id)
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, status, version, created_at;
	`
	err = repo.db.QueryRow(
		createQuery,
		name,
		description,
		serviceType,
		organizarionId,
		userId,
	).Scan(&id, &status, &version, &createdAt)

	if err != nil {
		return
	}

	tender = &model.Tender{
		Id:             id,
		Name:           name,
		Description:    description,
		ServiceType:    serviceType,
		Status:         status,
		OrganizationId: organizarionId,
		Version:        version,
		CreatedAt:      createdAt,
	}
	return
}

func (repo *TenderRepo) GetTendersByUserId(
	offset int,
	limit int,
	userId uuid.UUID,
) (tenders []*model.Tender, err error) {
	selectQuery := `
		SELECT id, name, description,
		service_type, status, organization_id,
		version, created_at
		FROM tender 
		WHERE user_id = $1
		ORDER BY name ASC
	`
	if limit > 0 {
		selectQuery += " LIMIT " + strconv.Itoa(limit)
	}

	if offset > 0 {
		selectQuery += " OFFSET " + strconv.Itoa(offset)
	}

	selectQuery += ";"

	rows, err := repo.db.Query(selectQuery, userId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tender model.Tender
		err = rows.Scan(
			&tender.Id,
			&tender.Name,
			&tender.Description,
			&tender.ServiceType,
			&tender.Status,
			&tender.OrganizationId,
			&tender.Version,
			&tender.CreatedAt,
		)
		if err != nil {
			return
		}
		tenders = append(tenders, &tender)
	}
	return
}

func (repo *TenderRepo) GetTenders(
	serviceTypeFlt string,
	offsetFlt string,
	limitFlt string,
	orgsIdFlt []uuid.UUID,
) (tenders []*model.Tender, err error) {
	whereClauses := []string{}
	if len(orgsIdFlt) != 0 {
		orgsIdStr := []string{}
		for _, orgID := range orgsIdFlt {
			orgsIdStr = append(orgsIdStr, "'"+orgID.String()+"'")
		}
		whereClauses = append(whereClauses, "organization_id IN ("+strings.Join(orgsIdStr, ",")+")")
	}

	if serviceTypeFlt != "" {
		whereClauses = append(whereClauses, "service_type = '"+serviceTypeFlt+"'")
	}

	selectQuery := `
		SELECT id, name, description,
		service_type, status, organization_id,
		version, created_at
		FROM tender 
	`
	if len(whereClauses) > 0 {
		selectQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	selectQuery += " ORDER BY name ASC "

	if limitFlt != "" {
		selectQuery += " LIMIT " + limitFlt
	}

	if offsetFlt != "" {
		selectQuery += " OFFSET " + offsetFlt
	}

	selectQuery += ";"

	rows, err := repo.db.Query(selectQuery)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tender model.Tender
		err = rows.Scan(
			&tender.Id,
			&tender.Name,
			&tender.Description,
			&tender.ServiceType,
			&tender.Status,
			&tender.OrganizationId,
			&tender.Version,
			&tender.CreatedAt,
		)
		if err != nil {
			return
		}
		tenders = append(tenders, &tender)
	}
	return
}

func (repo *TenderRepo) GetTenderById(id uuid.UUID) (tender *model.Tender, err error) {
	selectQuery := `
		SELECT id, name, description,
		service_type, status, organization_id,
		version, created_at
		FROM tender 
		WHERE id = $1
	`
	tender = &model.Tender{}
	err = repo.db.QueryRow(
		selectQuery, id,
	).Scan(
		&tender.Id,
		&tender.Name,
		&tender.Description,
		&tender.ServiceType,
		&tender.Status,
		&tender.OrganizationId,
		&tender.Version,
		&tender.CreatedAt,
	)

	return
}

func (repo *TenderRepo) UpdateTender(tenderUpd *model.Tender) (*model.Tender, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	selectOldQuery := `
		SELECT id, name, description, service_type,
		status, organization_id, version, created_at
		FROM tender 
		WHERE id = $1;
	`
	tenderOld := model.Tender{}
	err = tx.QueryRow(
		selectOldQuery, tenderUpd.Id,
	).Scan(
		&tenderOld.Id,
		&tenderOld.Name,
		&tenderOld.Description,
		&tenderOld.ServiceType,
		&tenderOld.Status,
		&tenderOld.OrganizationId,
		&tenderOld.Version,
		&tenderOld.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	createHistoryQuery := `
		INSERT INTO tender_history 
		(id, name, description, service_type,
		status, organization_id, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err = tx.Exec(
		createHistoryQuery,
		&tenderOld.Id,
		&tenderOld.Name,
		&tenderOld.Description,
		&tenderOld.ServiceType,
		&tenderOld.Status,
		&tenderOld.OrganizationId,
		&tenderOld.Version,
		&tenderOld.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	updateQuery := `
		UPDATE tender 
		SET name = $1, description = $2,
		service_type = $3, status = $4,
		organization_id = $5, version = $6
		WHERE id = $7;
	`
	tenderUpd.Version += 1
	_, err = tx.Exec(
		updateQuery,
		tenderUpd.Name,
		tenderUpd.Description,
		tenderUpd.ServiceType,
		tenderUpd.Status,
		tenderUpd.OrganizationId,
		tenderUpd.Version,
		tenderUpd.Id,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return tenderUpd, err
}

func (repo *TenderRepo) RollBackTender(id uuid.UUID, version int32) (*model.Tender, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	selectQuery := `
		SELECT id, name, description, service_type,
		status, organization_id, version, created_at
		FROM tender 
		WHERE id = $1;
	`
	tender := model.Tender{}
	err = tx.QueryRow(
		selectQuery, id,
	).Scan(
		&tender.Id,
		&tender.Name,
		&tender.Description,
		&tender.ServiceType,
		&tender.Status,
		&tender.OrganizationId,
		&tender.Version,
		&tender.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	createHistoryQuery := `
		INSERT INTO tender_history 
		(id, name, description, service_type,
		status, organization_id, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err = tx.Exec(
		createHistoryQuery,
		&tender.Id,
		&tender.Name,
		&tender.Description,
		&tender.ServiceType,
		&tender.Status,
		&tender.OrganizationId,
		&tender.Version,
		&tender.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	selectVersionQuery := `
		SELECT id, name, description, service_type,
		status, organization_id, version, created_at
		FROM tender_history
		WHERE id = $1 AND version = $2;
	`
	tenderOld := model.Tender{}
	err = tx.QueryRow(
		selectVersionQuery, id, version,
	).Scan(
		&tenderOld.Id,
		&tenderOld.Name,
		&tenderOld.Description,
		&tenderOld.ServiceType,
		&tenderOld.Status,
		&tenderOld.OrganizationId,
		&tenderOld.Version,
		&tenderOld.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	updateQuery := `
		UPDATE tender 
		SET name = $1, description = $2,
		service_type = $3, status = $4,
		organization_id = $5, version = $6
		WHERE id = $7;
	`
	tenderOld.Version = tender.Version + 1
	_, err = tx.Exec(
		updateQuery,
		&tenderOld.Name,
		&tenderOld.Description,
		&tenderOld.ServiceType,
		&tenderOld.Status,
		&tenderOld.OrganizationId,
		&tenderOld.Version,
		&tenderOld.Id,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &tenderOld, err
}

func NewRepo() (repo *TenderRepo, err error) {
	db, err := database.Connect()
	if err != nil {
		return
	}
	repo = &TenderRepo{db: db}

	table, err := repo.tableExists()
	if err != nil {
		return
	}

	if table {
		return
	}

	err = repo.createTable()
	if err == nil {
		slog.Info("Table 'bid' and 'bid_histrory' is created")
	} else {
		slog.Info("Can not create Table 'bid' and 'bid_histrory'")
	}
	return
}

func (repo *TenderRepo) tableExists() (table bool, err error) {
	rows, err := repo.db.Query(
		`SELECT EXISTS (SELECT FROM information_schema.tables 
		WHERE table_name = 'tender');`,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&table)
		if err != nil {
			return
		}
	}
	return
}

func (repo *TenderRepo) createTable() error {
	createTenderStatus := `
		CREATE TYPE tender_status 
		AS ENUM ('Created', 'Published', 'Closed');
	`
	createTenderServiceType := `
		CREATE TYPE tender_service_type 
		AS ENUM ('Construction', 'Delivery', 'Manufacture');
	`
	createTenderTable := `
		CREATE TABLE tender (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) UNIQUE NOT NULL,
		description VARCHAR(500) NOT NULL,
		service_type tender_service_type NOT NULL,
		status tender_status DEFAULT 'Created',
		user_id UUID REFERENCES employee(id),
		organization_id UUID REFERENCES organization(id),
		version INT DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	`
	createTenderHistoryTable := `
		CREATE TABLE tender_history (
		id UUID NOT NULL,
		name VARCHAR(100) NOT NULL,
		description VARCHAR(500) NOT NULL,
		service_type tender_service_type NOT NULL,
		status tender_status NOT NULL,
		organization_id UUID NOT NULL,
		version INT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		PRIMARY KEY (id, version));
	`
	_, err := repo.db.Exec(
		createTenderStatus +
			createTenderServiceType +
			createTenderTable +
			createTenderHistoryTable,
	)
	return err
}
