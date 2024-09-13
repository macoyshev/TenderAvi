package bid

import (
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"avi/internal/database"
	"avi/internal/model"

	"github.com/google/uuid"
)

type BidRepo struct {
	db *sql.DB
}

func (repo *BidRepo) CreateBid(
	name string,
	description string,
	tenderId uuid.UUID,
	authorType model.BidAuthorType,
	authorId uuid.UUID,
) (bid *model.Bid, err error) {
	var id uuid.UUID
	var version int32
	var createdAt time.Time
	var status model.BidStatus

	createQuery := `
		INSERT INTO bid 
		(name, description, tender_id, 
		author_type, author_id)
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, status, version, created_at;
	`
	err = repo.db.QueryRow(
		createQuery,
		name,
		description,
		tenderId,
		authorType,
		authorId,
	).Scan(&id, &status, &version, &createdAt)

	if err != nil {
		return
	}

	bid = &model.Bid{
		Id:          id,
		Name:        name,
		Description: description,
		AuthorType:  authorType,
		AuthorId:    authorId,
		Status:      status,
		TenderId:    tenderId,
		Version:     version,
		CreatedAt:   createdAt,
	}
	return
}

func (repo *BidRepo) GetBidsByAuthorId(
	offset int,
	limit int,
	authorId uuid.UUID,
) (bids []*model.Bid, err error) {
	selectQuery := `
		SELECT id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at
		FROM bid 
		WHERE author_id = $1
	`
	if limit > 0 {
		selectQuery += " LIMIT " + strconv.Itoa(limit)
	}

	if offset > 0 {
		selectQuery += " OFFSET " + strconv.Itoa(offset)
	}

	selectQuery += ";"

	rows, err := repo.db.Query(selectQuery, authorId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var bid model.Bid
		err = rows.Scan(
			&bid.Id,
			&bid.Name,
			&bid.Description,
			&bid.Status,
			&bid.TenderId,
			&bid.AuthorType,
			&bid.AuthorId,
			&bid.Version,
			&bid.CreatedAt,
		)
		if err != nil {
			return
		}
		bids = append(bids, &bid)
	}
	return
}

func (repo *BidRepo) GetBidsByTenderId(
	offset int,
	limit int,
	tenderId uuid.UUID,
) (bids []*model.Bid, err error) {
	selectQuery := `
		SELECT id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at
		FROM bid
		WHERE tender_id = $1
		ORDER BY name ASC
	`
	if limit > 0 {
		selectQuery += " LIMIT " + strconv.Itoa(limit)
	}

	if offset > 0 {
		selectQuery += " OFFSET " + strconv.Itoa(offset)
	}

	selectQuery += ";"

	rows, err := repo.db.Query(selectQuery, tenderId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var bid model.Bid
		err = rows.Scan(
			&bid.Id,
			&bid.Name,
			&bid.Description,
			&bid.Status,
			&bid.TenderId,
			&bid.AuthorType,
			&bid.AuthorId,
			&bid.Version,
			&bid.CreatedAt,
		)
		if err != nil {
			return
		}
		bids = append(bids, &bid)
	}
	return
}

func (repo *BidRepo) GetBidById(id uuid.UUID) (bid *model.Bid, err error) {
	selectQuery := `
		SELECT id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at
		FROM bid
		WHERE id = $1;
	`
	bid = &model.Bid{}
	err = repo.db.QueryRow(
		selectQuery, id,
	).Scan(
		&bid.Id,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderId,
		&bid.AuthorType,
		&bid.AuthorId,
		&bid.Version,
		&bid.CreatedAt,
	)
	return
}

func (repo *BidRepo) UpdateBidStatusById(
	id uuid.UUID, status model.BidStatus,
) (bid *model.Bid, err error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return
	}

	selectQuery := `
		SELECT id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at
		FROM bid
		WHERE id = $1;
	`
	bid = &model.Bid{}
	err = tx.QueryRow(
		selectQuery, id,
	).Scan(
		&bid.Id,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderId,
		&bid.AuthorType,
		&bid.AuthorId,
		&bid.Version,
		&bid.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	createQuery := `
		INSERT INTO bid_history
		(id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`
	_, err = tx.Exec(
		createQuery,
		bid.Id,
		bid.Name,
		bid.Description,
		bid.Status,
		bid.TenderId,
		bid.AuthorType,
		bid.AuthorId,
		bid.Version,
		bid.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	updateQuery := `
		UPDATE bid 
		SET status = $1, version = $2
		WHERE id = $3;
	`
	bid.Version += 1
	bid.Status = status
	_, err = tx.Exec(updateQuery, status, bid.Version, id)
	if err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	return
}

func (repo *BidRepo) EditBidById(
	id uuid.UUID, name string, description string,
) (bid *model.Bid, err error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return
	}

	selectQuery := `
		SELECT id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at
		FROM bid
		WHERE id = $1;
	`
	bid = &model.Bid{}
	err = tx.QueryRow(
		selectQuery, id,
	).Scan(
		&bid.Id,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderId,
		&bid.AuthorType,
		&bid.AuthorId,
		&bid.Version,
		&bid.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	createQuery := `
		INSERT INTO bid_history
		(id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`
	_, err = tx.Exec(
		createQuery,
		bid.Id,
		bid.Name,
		bid.Description,
		bid.Status,
		bid.TenderId,
		bid.AuthorType,
		bid.AuthorId,
		bid.Version,
		bid.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	updateQuery := `
		UPDATE bid 
		SET name = $1, description = $2, version = $3
		WHERE id = $4;
	`
	bid.Version += 1
	bid.Name = name
	bid.Description = description
	_, err = tx.Exec(
		updateQuery, bid.Name, bid.Description, bid.Version, id,
	)
	if err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	return
}

func (repo *BidRepo) CreateReviewById(
	id uuid.UUID, description string,
) (bid *model.Bid, err error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return
	}

	selectQuery := `
		SELECT id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at
		FROM bid
		WHERE id = $1;
	`
	bid = &model.Bid{}
	err = tx.QueryRow(
		selectQuery, id,
	).Scan(
		&bid.Id,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderId,
		&bid.AuthorType,
		&bid.AuthorId,
		&bid.Version,
		&bid.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	createQuery := `
		INSERT INTO review
		(description, bid_id)
		VALUES ($1, $2);
	`
	_, err = tx.Exec(createQuery, description, id)

	if err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	return
}

func (repo *BidRepo) RollbackById(
	id uuid.UUID, version int32,
) (bid *model.Bid, err error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return
	}

	selectQuery := `
		SELECT id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at
		FROM bid
		WHERE id = $1;
	`
	bid = &model.Bid{}
	err = tx.QueryRow(
		selectQuery, id,
	).Scan(
		&bid.Id,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderId,
		&bid.AuthorType,
		&bid.AuthorId,
		&bid.Version,
		&bid.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	createQuery := `
		INSERT INTO bid_history
		(id, name, description,
		status, tender_id, author_type,
		author_id, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`
	_, err = tx.Exec(
		createQuery,
		bid.Id,
		bid.Name,
		bid.Description,
		bid.Status,
		bid.TenderId,
		bid.AuthorType,
		bid.AuthorId,
		bid.Version,
		bid.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	selectVersionQuery := `
		SELECT name, description, status, 
		tender_id, author_type, author_id
		FROM bid_history
		WHERE id = $1 AND version = $2;
	`
	bid.Version += 1
	err = tx.QueryRow(
		selectVersionQuery, id, version,
	).Scan(
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderId,
		&bid.AuthorType,
		&bid.AuthorId,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	updateQuery := `
		UPDATE bid 
		SET name = $1, description = $2, 
		status = $3, tender_id = $4, author_type = $5,
		author_id = $6, version = $7
		WHERE id = $8;
	`
	_, err = tx.Exec(
		updateQuery,
		bid.Name,
		bid.Description,
		bid.Status,
		bid.TenderId,
		bid.AuthorType,
		bid.AuthorId,
		bid.Version,
		id,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	return
}

func (repo *BidRepo) GetReviews(
	offset int,
	limit int,
	authorId uuid.UUID,
	tenderId uuid.UUID,
) (reviews []*model.Review, err error) {
	selectQuery := `
		SELECT review.id, review.description, review.created_at 
		FROM bid 
		INNER JOIN review on bid.id = review.bid_id
		WHERE bid.tender_id = $1 and bid.author_id=$2
	`
	if limit > 0 {
		selectQuery += " LIMIT " + strconv.Itoa(limit)
	}

	if offset > 0 {
		selectQuery += " OFFSET " + strconv.Itoa(offset)
	}

	selectQuery += ";"

	rows, err := repo.db.Query(selectQuery, tenderId, authorId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var review model.Review
		err = rows.Scan(
			&review.Id,
			&review.Description,
			&review.CreatedAt,
		)
		if err != nil {
			return
		}
		reviews = append(reviews, &review)
	}
	return
}

func (repo *BidRepo) GetDisicions(id uuid.UUID) (rejects int, approves int, err error) {
	selectQuery := `
		SELECT rejects, approves
		FROM bid
		WHERE id = $1;
	`
	err = repo.db.QueryRow(selectQuery, id).Scan(&rejects, &approves)
	return
}

func (repo *BidRepo) UpdateApproves(
	bidId uuid.UUID, approves int, tenderId uuid.UUID, closeTender bool,
) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	updateQuery := `
		UPDATE bid 
		SET approves = $1
		WHERE id = $2;
	`
	_, err = tx.Exec(updateQuery, approves, bidId)
	if err != nil {
		tx.Rollback()
		return err
	}

	if closeTender {
		updateQuery = `
			UPDATE tender 
			SET status = 'Closed'
			WHERE id = $1;
		`
		_, err = tx.Exec(updateQuery, tenderId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	return nil
}

func (repo *BidRepo) UpdateRejects(bidId uuid.UUID, rejects int) error {
	updateQuery := `
		UPDATE bid 
		SET rejects = $1
		WHERE id = $2;
	`
	_, err := repo.db.Exec(updateQuery, rejects, bidId)
	return err
}

func NewRepo() (repo *BidRepo, err error) {
	db, err := database.Connect()
	if err != nil {
		return
	}
	repo = &BidRepo{db: db}

	table, err := repo.tableExists()
	if err != nil {
		return
	}

	if table {
		return
	}

	err = repo.createTable()
	if err == nil {
		slog.Info("Tables 'bid', 'bid_histrory' and 'review' are created")
	} else {
		slog.Info("Can not create tables 'bid', 'bid_histrory' and 'review'")
	}

	return
}

func (repo *BidRepo) tableExists() (table bool, err error) {
	rows, err := repo.db.Query(
		`SELECT EXISTS (SELECT FROM information_schema.tables 
		WHERE table_name = 'bid');`,
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

func (repo *BidRepo) createTable() error {
	createBidStatus := `
		CREATE TYPE bid_status 
		AS ENUM ('Created', 'Published', 'Canceled');
	`
	createBidAuthorType := `
		CREATE TYPE bid_author_type 
		AS ENUM ('User', 'Organization');
	`
	createBidTable := `
		CREATE TABLE bid (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) UNIQUE NOT NULL,
		description VARCHAR(500) NOT NULL,
		status bid_status DEFAULT 'Created',
		tender_id UUID REFERENCES tender(id),
		author_type bid_author_type NOT NULL,
		author_id UUID NOT NULL,
		version INT DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		rejects INT DEFAULT 0,
		approves INT DEFAULT 0);
	`
	createBidHistoryTable := `
		CREATE TABLE bid_history (
		id UUID NOT NULL,
		name VARCHAR(100) NOT NULL,
		description VARCHAR(500) NOT NULL,
		status bid_status NOT NULL,
		tender_id UUID NOT NULL,
		author_type bid_author_type NOT NULL,
		author_id UUID NOT NULL,
		version INT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		PRIMARY KEY (id, version));
	`
	createReviewTable := `
		CREATE TABLE review (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		description VARCHAR(1000) NOT NULL,
		bid_id UUID REFERENCES bid(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	`
	_, err := repo.db.Exec(
		createBidAuthorType +
			createBidStatus +
			createBidHistoryTable +
			createBidTable +
			createReviewTable,
	)
	return err
}
