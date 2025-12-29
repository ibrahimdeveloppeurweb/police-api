package officers

// OfficerDashboardResponse represents the dashboard response for mobile app
// Maps to Flutter's OfficerDashboardStats model
type OfficerDashboardResponse struct {
	Controls      PeriodStats       `json:"controls"`
	Inspections   PeriodStats       `json:"inspections"`
	Revenues      RevenueStats      `json:"revenues"`
	Compliance    ComplianceStats   `json:"compliance"`
	Infractions   []InfractionRank  `json:"infractions"`
	ActivityChart []int             `json:"activity_chart"`
	Period        string            `json:"period"`
}

// PeriodStats represents statistics by time period (today, week, month, year)
type PeriodStats struct {
	Total int `json:"total"`
	Today int `json:"today"`
	Week  int `json:"week"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

// RevenueStats represents revenue statistics by time period
type RevenueStats struct {
	Total float64 `json:"total"`
	Today float64 `json:"today"`
	Week  float64 `json:"week"`
	Month float64 `json:"month"`
	Year  float64 `json:"year"`
}

// ComplianceStats represents compliance statistics
type ComplianceStats struct {
	VehiclesConformed int     `json:"vehicles_conformed"`
	VehiclesInspected int     `json:"vehicles_inspected"`
	ConformanceRate   float64 `json:"conformance_rate"`
}

// InfractionRank represents a ranked infraction type
type InfractionRank struct {
	Name  string `json:"name"`
	Code  string `json:"code"`
	Count int    `json:"count"`
	Rank  int    `json:"rank"`
}

// OfficerStatisticsResponse represents officer statistics for mobile app
// Maps to Flutter's OfficerStatistics model
type OfficerStatisticsResponse struct {
	TotalIssuedTickets int     `json:"total_issued_tickets"`
	ActiveTickets      int     `json:"active_tickets"`
	TotalPayments      int     `json:"total_payments"`
	TotalFineAmount    float64 `json:"total_fine_amount"`
}
