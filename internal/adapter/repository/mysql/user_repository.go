package mysql

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

type UserModel struct {
    ID          int64  `gorm:"primaryKey;autoIncrement"`
    Email       string `gorm:"column:email;uniqueIndex:uk_users_email;size:255;not null"`
    DisplayName string `gorm:"column:display_name;size:255;not null"`
    PictureURL  string `gorm:"column:picture_url;size:512"`
}

func (UserModel) TableName() string { return "users" }

type RoleModel struct {
    ID          int64  `gorm:"primaryKey;autoIncrement"`
    Name        string `gorm:"column:name;size:64;uniqueIndex:uk_roles_name;not null"`
    Description string `gorm:"column:description;size:255"`
}

func (RoleModel) TableName() string { return "roles" }

type UserRoleModel struct {
    UserID int64 `gorm:"column:user_id;primaryKey"`
    RoleID int64 `gorm:"column:role_id;primaryKey"`
}

func (UserRoleModel) TableName() string { return "user_roles" }

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*usecase.User, error) {
    var u UserModel
    if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    roles, err := r.loadRoles(ctx, id)
    if err != nil { return nil, err }
    return &usecase.User{
        ID: u.ID,
        Email: u.Email,
        DisplayName: u.DisplayName,
        PictureURL: u.PictureURL,
        Roles: roles,
    }, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, displayName, pictureURL string) (*usecase.User, error) {
    if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Updates(map[string]any{
        "display_name": displayName,
        "picture_url":  pictureURL,
    }).Error; err != nil {
        return nil, err
    }
    return r.FindByID(ctx, id)
}

func (r *UserRepository) loadRoles(ctx context.Context, userID int64) ([]string, error) {
    type row struct{ Name string }
    var rows []row
    q := r.db.WithContext(ctx).Table("user_roles ur").
        Joins("JOIN roles r ON r.id = ur.role_id").
        Where("ur.user_id = ?", userID).
        Select("r.name as name")
    if err := q.Scan(&rows).Error; err != nil { return nil, err }
    out := make([]string, 0, len(rows))
    for _, rr := range rows { out = append(out, rr.Name) }
    return out, nil
}

