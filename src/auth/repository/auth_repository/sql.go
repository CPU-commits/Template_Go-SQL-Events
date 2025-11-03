package auth_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/model"
	"github.com/CPU-commits/Template_Go-EventDriven/src/package/db/models"
	"github.com/CPU-commits/Template_Go-EventDriven/src/utils"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	. "github.com/aarondl/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

type sqlAuthRepository struct {
	db *sql.DB
}

func (sqlAuthRepository) sqlAuthToAuth(auth *models.Auth) *model.Auth {
	return &model.Auth{
		ID:         auth.ID,
		Password:   auth.Password.String,
		IDUser:     auth.IDUser,
		Provider:   auth.Provider,
		ProviderId: auth.ProviderId.String,
	}
}

func (sqlAuthRepository) criteriaToWhere(criteria *Criteria) []QueryMod {
	if criteria == nil {
		return nil
	}
	where := []QueryMod{}
	if criteria.Provider != "" {
		where = append(where, models.AuthWhere.Provider.EQ(criteria.Provider))
	}
	if criteria.IDUser != 0 {
		where = append(where, models.AuthWhere.IDUser.EQ(criteria.IDUser))
	}

	return where
}

func (sqlAR sqlAuthRepository) Exists(criteria *Criteria) (bool, error) {
	where := sqlAR.criteriaToWhere(criteria)

	exists, err := models.Auths(where...).Exists(context.Background(), sqlAR.db)
	if err != nil {
		return false, utils.ErrRepositoryFailed
	}

	return exists, nil
}

func (sqlAR sqlAuthRepository) InsertOne(auth *model.Auth) (int64, error) {
	var password *string
	if auth.Password != "" {
		passwordHashed, err := bcrypt.GenerateFromPassword([]byte(auth.Password), bcrypt.DefaultCost)
		if err != nil {
			return 0, err
		}
		passwordStr := string(passwordHashed)
		password = &passwordStr
	}

	sqlAuth := models.Auth{
		IDUser:     auth.IDUser,
		Provider:   auth.Provider,
		ProviderId: null.NewString(auth.ProviderId, auth.ProviderId != ""),
		Password:   null.StringFromPtr(password),
	}

	if err := sqlAuth.Insert(context.Background(), sqlAR.db, boil.Infer()); err != nil {
		fmt.Printf("err: %v\n", err)
		return 0, utils.ErrRepositoryFailed
	}

	return sqlAuth.ID, nil
}

func (sqlAR sqlAuthRepository) FindOneByUsernameAndProvider(username string, provider string) (*model.Auth, error) {
	auth, err := models.Auths(
		InnerJoin("users u on u.id = auths.id_user"),
		Where("u.email = ?", username),
		models.AuthWhere.Provider.EQ(provider),
	).One(context.Background(), sqlAR.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, utils.ErrRepositoryFailed
	}

	return sqlAR.sqlAuthToAuth(auth), nil
}

func (sqlAR sqlAuthRepository) FindOneByUserIdAndProvider(userId int64, provider string) (*model.Auth, error) {
	auth, err := models.Auths(
		Where("id_user = ?", userId),
		models.AuthWhere.Provider.EQ(provider),
	).One(context.Background(), sqlAR.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, utils.ErrRepositoryFailed
	}

	return sqlAR.sqlAuthToAuth(auth), nil
}

func (sqlAR sqlAuthRepository) FindOneByUserId(userId int64) (*model.Auth, error) {
	auth, err := models.Auths(
		Where("id_user = ?", userId),
	).One(context.Background(), sqlAR.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, utils.ErrRepositoryFailed
	}

	return sqlAR.sqlAuthToAuth(auth), nil
}

func (sqlAR sqlAuthRepository) UpdatePassword(userId int64, data DataUpdate) error {
	auth, err := models.Auths(
		models.AuthWhere.IDUser.EQ(userId),
		models.AuthWhere.Provider.EQ("local"),
	).One(context.Background(), sqlAR.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return utils.ErrRepositoryFailed
	}

	var cols []string
	if data.Password != nil {
		passwordHashed, err := bcrypt.GenerateFromPassword([]byte(*data.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		auth.Password = null.StringFrom(string(passwordHashed))
		cols = append(cols, models.AuthColumns.Password)
	}

	if _, err := auth.Update(context.Background(), sqlAR.db, boil.Whitelist(cols...)); err != nil {
		return utils.ErrRepositoryFailed
	}
	return nil
}

func NewSQLAuthRepository(db *sql.DB) AuthRepository {
	return sqlAuthRepository{
		db: db,
	}
}
