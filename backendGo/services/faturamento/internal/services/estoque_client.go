package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
)

type EstoqueProdutoPreco struct {
	PrecoVista float64 `json:"preco_vista"`
	PrecoPrazo float64 `json:"preco_prazo"`
	VigenteEm  string  `json:"vigente_em"`
}

type EstoqueProduto struct {
	ID          uint64              `json:"id"`
	Codigo      string              `json:"codigo"`
	Descricao   string              `json:"descricao"`
	Saldo       float64             `json:"saldo"`
	PrecoAtual  *EstoqueProdutoPreco `json:"preco_atual"`
}

type BaixaEstoqueInput struct {
	ProdutoID      uint64  `json:"produto_id"`
	Tipo           string  `json:"tipo"`
	Quantidade     float64 `json:"quantidade"`
	Origem         string  `json:"origem"`
	Referencia     string  `json:"referencia"`
	IdempotencyKey string  `json:"idempotency_key"`
}

type EstoqueCliente interface {
	BuscarProduto(produtoID uint64) (*EstoqueProduto, error)
	SolicitarBaixa(input BaixaEstoqueInput) error
}

type estoqueCliente struct {
	baseURL string
	client  *http.Client
}

func NewEstoqueCliente(baseURL string) EstoqueCliente {
	return &estoqueCliente{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *estoqueCliente) BuscarProduto(produtoID uint64) (*EstoqueProduto, error) {
	url := fmt.Sprintf("%s/produtos/%d", c.baseURL, produtoID)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("Serviço de Estoque indisponível")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var produto EstoqueProduto
		if err := json.NewDecoder(resp.Body).Decode(&produto); err != nil {
			return nil, apperrors.Internal("Resposta inválida do Serviço de Estoque")
		}
		return &produto, nil
	case http.StatusNotFound:
		return nil, apperrors.NotFound("Produto não encontrado no Serviço de Estoque")
	default:
		return nil, apperrors.ServiceUnavailable("Não foi possível consultar o Serviço de Estoque")
	}
}

type estoqueErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *estoqueCliente) SolicitarBaixa(input BaixaEstoqueInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return apperrors.Internal("Erro ao montar solicitação de baixa")
	}

	url := fmt.Sprintf("%s/estoque/movimentos", c.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return apperrors.Internal("Erro ao montar solicitação de baixa")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return apperrors.ServiceUnavailable("Serviço de Estoque indisponível")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}

	var envelope estoqueErrorEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&envelope)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return apperrors.NotFound("Produto não encontrado no Serviço de Estoque")
	case http.StatusConflict:
		msg := envelope.Error.Message
		if msg == "" {
			msg = "Saldo insuficiente para a baixa no Estoque"
		}
		return apperrors.Conflict(msg)
	case http.StatusUnprocessableEntity:
		msg := envelope.Error.Message
		if msg == "" {
			msg = "Baixa inválida no Serviço de Estoque"
		}
		return apperrors.Unprocessable(msg)
	case http.StatusBadRequest:
		msg := envelope.Error.Message
		if msg == "" {
			msg = "Solicitação de baixa inválida no Serviço de Estoque"
		}
		return apperrors.Internal(msg)
	default:
		return apperrors.ServiceUnavailable("Não foi possível processar a baixa no Serviço de Estoque")
	}
}
