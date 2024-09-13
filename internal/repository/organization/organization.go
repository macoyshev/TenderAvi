package organization

import (
	"database/sql"

	"avi/internal/database"
	"avi/internal/model"

	"github.com/google/uuid"
)

type OrganizationRepo struct {
	db *sql.DB
}

func (repo *OrganizationRepo) GetOrganizationById(
	id uuid.UUID,
) (organization *model.Organization, err error) {
	selectQuery := `
		SELECT id, name, description, type, 
		created_at, updated_at
		FROM organization WHERE id = $1;
	`
	organization = &model.Organization{}
	row := repo.db.QueryRow(selectQuery, id)
	err = row.Scan(
		&organization.Id,
		&organization.Name,
		&organization.Description,
		&organization.OrganizationType,
		&organization.CreatedAt,
		&organization.UpdatedAt,
	)
	return
}

func (repo *OrganizationRepo) GetOrganizationsByUserId(
	user_id uuid.UUID,
) (orgs []*model.Organization, err error) {
	selectQuery := `
		SELECT organization.id, organization.name, 
		organization.description, organization.type, 
		organization.created_at, organization.updated_at
		FROM organization_responsible
		INNER JOIN organization 
		ON organization.id = organization_responsible.organization_id
		WHERE organization_responsible.user_id = $1;
	`
	rows, err := repo.db.Query(selectQuery, user_id)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var org model.Organization
		err = rows.Scan(
			&org.Id,
			&org.Name,
			&org.Description,
			&org.OrganizationType,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return
		}
		orgs = append(orgs, &org)
	}
	return
}

func (repo *OrganizationRepo) GetResponsibleUsersId(orgId uuid.UUID) ([]uuid.UUID, error) {
	selectQuery := `
		SELECT user_id 
		FROM organization_responsible
		WHERE organization_id = $1;
	`
	rows, err := repo.db.Query(selectQuery, orgId)
	ids := []uuid.UUID{}
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func NewRepo() (repo *OrganizationRepo, err error) {
	db, err := database.Connect()
	if err != nil {
		return
	}
	repo = &OrganizationRepo{db: db}
	return
}
