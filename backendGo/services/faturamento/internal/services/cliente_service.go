package services

import (
	"errors"
	"strings"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
	"github.com/korp-teste/backendGo/services/faturamento/internal/repositories"
)

type CreateClienteInput struct {
	Nome string `json:"nome"`
}

type UpdateClienteInput struct {
	Nome string `json:"nome"`
}

type ClienteService interface {
	Create(input CreateClienteInput) (*models.Cliente, error)
	Get(id uint64) (*models.Cliente, error)
	List() ([]models.Cliente, error)
	Update(id uint64, input UpdateClienteInput) (*models.Cliente, error)
	Delete(id uint64) error
}

type clienteService struct {
	repo repositories.ClienteRepository
}

func NewClienteService(repo repositories.ClienteRepository) ClienteService {
	return &clienteService{repo: repo}
}

func (s *clienteService) Create(input CreateClienteInput) (*models.Cliente, error) {
	input.Nome = strings.TrimSpace(input.Nome)
	if input.Nome == "" {
		return nil, apperrors.BadRequest("Nome é obrigatório")
	}
	c := &models.Cliente{Nome: input.Nome}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return s.repo.FindByID(c.ID)
}

func (s *clienteService) Get(id uint64) (*models.Cliente, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Cliente não encontrado")
		}
		return nil, err
	}
	return c, nil
}

func (s *clienteService) List() ([]models.Cliente, error) {
	return s.repo.List()
}

func (s *clienteService) Update(id uint64, input UpdateClienteInput) (*models.Cliente, error) {
	input.Nome = strings.TrimSpace(input.Nome)
	if input.Nome == "" {
		return nil, apperrors.BadRequest("Nome é obrigatório")
	}
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Cliente não encontrado")
		}
		return nil, err
	}
	c.Nome = input.Nome
	if err := s.repo.Update(c); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *clienteService) Delete(id uint64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return apperrors.NotFound("Cliente não encontrado")
		}
		return err
	}
	tem, err := s.repo.TemNotas(id)
	if err != nil {
		return err
	}
	if tem {
		return apperrors.Conflict("Não é possível excluir cliente com notas fiscais vinculadas")
	}
	return s.repo.Delete(id)
}
