package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/JullMol/titan-ledger/internal/core/ports"
)

type TitanHandler struct {
	walletService    ports.WalletService
	transferService  ports.TransactionService
}

func NewTitanHandler(w ports.WalletService, t ports.TransactionService) *TitanHandler {
	return &TitanHandler{
		walletService:   w,
		transferService: t,
	}
}

// CreateWallet menangani POST /wallets
func (h *TitanHandler) CreateWallet(c *fiber.Ctx) error {
	type Request struct {
		UserID string `json:"user_id"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	wallet, err := h.walletService.CreateWallet(c.Context(), req.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(wallet)
}

// GetBalance menangani GET /wallets/:id
func (h *TitanHandler) GetBalance(c *fiber.Ctx) error {
	id := c.Params("id")
	
	wallet, err := h.walletService.GetWalletBalance(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(wallet)
}

// Transfer menangani POST /transfer
func (h *TitanHandler) Transfer(c *fiber.Ctx) error {
	type TransferPayload struct {
		FromWalletID string `json:"from_wallet_id"`
		ToWalletID   string `json:"to_wallet_id"`
		Amount       int64  `json:"amount"`
		ReferenceID  string `json:"reference_id"`
	}

	var p TransferPayload
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	req := ports.TransferRequest{
		FromWalletID: p.FromWalletID,
		ToWalletID:   p.ToWalletID,
		Amount:       p.Amount,
		ReferenceID:  p.ReferenceID,
	}

	err := h.transferService.Transfer(c.Context(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Transfer processed"})
}