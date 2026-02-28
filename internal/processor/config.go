package processor

import "TrainTicketsTool/internal/invoice"

type Config struct {
	InputDir  string
	OutputDir string
	DateField invoice.DateField
}

