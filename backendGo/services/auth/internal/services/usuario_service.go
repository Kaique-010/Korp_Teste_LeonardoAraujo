package services

import (
	"context"
	"errors"
	"strings"

	"korp-teste/auth/internal/models"
	"korp-teste/auth/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailObrigatorio   = errors.New("email é obrigatório")
	ErrSenhaObrigatoria   = errors.New("senha é obrigatória")
	ErrUsuarioInativo     = errors.New("usuário inativo")
	ErrCredencialInvalida = errors.New("credencial inválida")
)

type UsuarioService struct {
	repo repositories.UsuarioRepository
}

func NewUsuarioService(repo repositories.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repo: repo}
}

func (s *UsuarioService) Criar(
	ctx context.Context,
	nome string,
	email string,
	senha string,
) (*models.Usuario, error) {

	email = strings.TrimSpace(strings.ToLower(email))
	nome = strings.TrimSpace(nome)

	if email == "" {
		return nil, ErrEmailObrigatorio
	}

	if senha == "" {
		return nil, ErrSenhaObrigatoria
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(senha),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	usuario := &models.Usuario{
		Nome:      nome,
		Email:     email,
		SenhaHash: string(hash),
		Ativo:     true,
	}

	if err := s.repo.Criar(ctx, usuario); err != nil {
		return nil, err
	}
	return usuario, nil
}

func (s *UsuarioService) Autenticar(
	ctx context.Context,
	email string,
	senha string,
) (*models.Usuario, error) {

	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return nil, ErrEmailObrigatorio
	}

	if senha == "" {
		return nil, ErrSenhaObrigatoria
	}

	usuario, err := s.repo.BuscarPorEmail(ctx, email)
	if err != nil {
		return nil, ErrCredencialInvalida
	}

	if !usuario.Ativo {
		return nil, ErrUsuarioInativo
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(usuario.SenhaHash),
		[]byte(senha),
	)

	if err != nil {
		return nil, ErrCredencialInvalida
	}

	return usuario, nil
}

func (s *UsuarioService) CriarSeVazio(
	ctx context.Context,
	nome string,
	email string,
	senha string,
) (bool, *models.Usuario, error) {
	total, err := s.repo.Contar(ctx)
	if err != nil {
		return false, nil, err
	}
	if total > 0 {
		return false, nil, nil
	}

	u, err := s.Criar(ctx, nome, email, senha)
	return true, u, err
}
