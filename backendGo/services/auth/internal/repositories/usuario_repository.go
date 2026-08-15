package repositories

import (
	"context"
	"korp-teste/auth/internal/models"

	"gorm.io/gorm"
)

type UsuarioRepository interface {
	Criar(ctx context.Context, usuario *models.Usuario) error
	BuscarPorEmail(ctx context.Context, email string) (*models.Usuario, error)
	BuscarPorID(ctx context.Context, id uint) (*models.Usuario, error)
	Atualizar(ctx context.Context, usuario *models.Usuario) error
	Contar(ctx context.Context) (int64, error)
}

type usuarioRepository struct {
	db *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) UsuarioRepository {
	return &usuarioRepository{db: db}
}

func (r *usuarioRepository) Criar(ctx context.Context, usuario *models.Usuario) error {
	return r.db.WithContext(ctx).Create(usuario).Error
}

func (r *usuarioRepository) BuscarPorEmail(ctx context.Context, email string) (*models.Usuario, error) {
	var usuario models.Usuario
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&usuario).Error
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

func (r *usuarioRepository) BuscarPorID(ctx context.Context, id uint) (*models.Usuario, error) {
	var usuario models.Usuario
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&usuario).Error
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

func (r *usuarioRepository) Atualizar(ctx context.Context, usuario *models.Usuario) error {
	return r.db.WithContext(ctx).Model(&models.Usuario{}).Where("id = ?", usuario.ID).Updates(usuario).Error
}

func (r *usuarioRepository) Contar(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Usuario{}).Count(&total).Error
	return total, err
}
