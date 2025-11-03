package user_repository

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

type sqlUserRepository struct {
	db *sql.DB
}

type SqlUserRepository = sqlUserRepository

func (sqlUserRepository) SqlUserToUser(
	sqlUser *models.User,
	roles []string,
) *model.User {
	if sqlUser.R != nil && sqlUser.R.IDUserRolesUsers != nil && roles == nil {
		sqlRoles := sqlUser.R.IDUserRolesUsers

		for _, sqlRole := range sqlRoles {
			roles = append(roles, sqlRole.Role.String())
		}
	}

	return &model.User{
		ID:    sqlUser.ID,
		Email: sqlUser.Email,
		Name:  sqlUser.Name,
		Roles: utils.MapNoError(roles, func(role string) model.Role {
			return model.Role(role)
		}),
		UpdatedAt: sqlUser.UpdatedAt,
	}
}

func (sqlUR sqlUserRepository) getUserRoles(user *models.User) ([]string, error) {
	roles, err := user.IDUserRolesUsers().All(context.Background(), sqlUR.db)
	if err != nil {
		return nil, utils.ErrRepositoryFailed
	}

	return utils.MapNoError(roles, func(role *models.RolesUser) string {
		return role.Role.String()
	}), nil
}

func (sqlUR sqlUserRepository) criteriaToWhere(criteria *Criteria) []QueryMod {
	var mod []QueryMod
	if criteria == nil {
		return nil
	}
	if criteria.ID != 0 {
		mod = append(mod, models.UserWhere.ID.EQ(criteria.ID))
	}
	if criteria.Email.EQ != nil {
		mod = append(mod, models.UserWhere.Email.EQ(*criteria.Email.EQ))
	} else if criteria.Email.IContains != nil {
		mod = append(mod, models.UserWhere.Email.ILIKE(fmt.Sprintf("%%%s%%", *criteria.Email.IContains)))
	}
	if criteria.Name.EQ != nil {
		mod = append(mod, models.UserWhere.Name.EQ(*criteria.Name.EQ))
	} else if criteria.Name.IContains != nil {
		mod = append(mod, models.UserWhere.Name.ILIKE(fmt.Sprintf("%%%s%%", *criteria.Name.IContains)))
	}
	if criteria.ID_NIN != nil {
		mod = append(mod, models.UserWhere.ID.NIN(criteria.ID_NIN))
	}
	if criteria.Roles != nil {
		mod = append(mod,
			Select(`"users"."id" AS "id"`),
			Load(models.UserRels.IDUserRolesUsers),
			LeftOuterJoin("roles_users ru on ru.id_user = users.id"),
			WhereIn("ru.role in ?", utils.MapNoError(criteria.Roles, func(role model.Role) any {
				return string(role)
			})...),
		)
	}

	var orMods []QueryMod
	for _, clause := range criteria.Or {
		orWhere := sqlUR.criteriaToWhere(&clause)
		if orWhere != nil {
			orMods = append(orMods, Or2(Expr(orWhere...)))
		}
	}
	if orMods != nil {
		mod = append(mod, Expr(orMods...))
	}
	return mod
}

func (sqlUR sqlUserRepository) Exists(criteria *Criteria) (bool, error) {
	where := sqlUR.criteriaToWhere(criteria)

	exists, err := models.Users(where...).Exists(context.Background(), sqlUR.db)
	if err != nil {
		return false, utils.ErrRepositoryFailed
	}

	return exists, nil
}

func (sqlUR sqlUserRepository) FindOneByEmail(email string) (*model.User, error) {
	user, err := models.Users(
		Where("email = ?", email),
	).One(context.Background(), sqlUR.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, utils.ErrRepositoryFailed
	}
	roles, err := sqlUR.getUserRoles(user)
	if err != nil {
		return nil, err
	}

	return sqlUR.SqlUserToUser(user, roles), nil
}

func (sqlUR sqlUserRepository) FindOneByID(id int64) (*model.User, error) {
	user, err := models.Users(
		Where("id = ?", id),
	).One(context.Background(), sqlUR.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, utils.ErrRepositoryFailed
	}
	roles, err := sqlUR.getUserRoles(user)
	if err != nil {
		return nil, err
	}

	return sqlUR.SqlUserToUser(user, roles), nil
}

func (sqlUR sqlUserRepository) InsertOne(user *model.User, auth *model.Auth) (*model.User, error) {
	sqlUser := models.User{
		Email: user.Email,
		Name:  user.Name,
	}
	ctx := context.Background()
	tx, err := sqlUR.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.ErrRepositoryFailed
	}
	if err := sqlUser.Insert(ctx, tx, boil.Infer()); err != nil {
		tx.Rollback()
		return nil, utils.ErrRepositoryFailed
	}
	for _, role := range user.Roles {
		sqlRole := models.RolesUser{
			IDUser: sqlUser.ID,
			Role:   models.RoleName(role),
		}
		if err := sqlRole.Insert(ctx, tx, boil.Infer()); err != nil {
			tx.Rollback()
			return nil, utils.ErrRepositoryFailed
		}
	}
	if auth != nil {
		var password *string
		if auth.Password != "" {
			passwordHashed, err := bcrypt.GenerateFromPassword([]byte(auth.Password), bcrypt.DefaultCost)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			passwordStr := string(passwordHashed)
			password = &passwordStr
		}

		sqlAuth := models.Auth{
			IDUser:     sqlUser.ID,
			Password:   null.StringFromPtr(password),
			Provider:   auth.Provider,
			ProviderId: null.NewString(auth.ProviderId, auth.ProviderId != ""),
		}
		if err := sqlAuth.Insert(ctx, tx, boil.Infer()); err != nil {
			tx.Rollback()

			return nil, utils.ErrRepositoryFailed
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, utils.ErrRepositoryFailed
	}

	return sqlUR.SqlUserToUser(&sqlUser, utils.MapNoError(user.Roles, func(role model.Role) string {
		return string(role)
	})), nil
}

func NewSQLUserRepository(db *sql.DB) UserRepository {
	return sqlUserRepository{
		db: db,
	}
}

func SqlExplicitUserRepository(db *sql.DB) SqlUserRepository {
	return sqlUserRepository{
		db: db,
	}
}
