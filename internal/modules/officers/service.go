package officers

import (
	"context"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/controle"
	"police-trafic-api-frontend-aligned/ent/inspection"
	"police-trafic-api-frontend-aligned/ent/procesverbal"
	"police-trafic-api-frontend-aligned/ent/user"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines officers service interface for mobile app
type Service interface {
	GetOfficerDashboard(ctx context.Context, officerID string, period string) (*OfficerDashboardResponse, error)
	GetOfficerStatistics(ctx context.Context, officerID string) (*OfficerStatisticsResponse, error)
}

type service struct {
	client         *ent.Client
	userRepo       repository.UserRepository
	infractionRepo repository.InfractionRepository
	logger         *zap.Logger
}

// NewService creates a new officers service
func NewService(
	client *ent.Client,
	userRepo repository.UserRepository,
	infractionRepo repository.InfractionRepository,
	logger *zap.Logger,
) Service {
	return &service{
		client:         client,
		userRepo:       userRepo,
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// GetOfficerDashboard returns dashboard statistics for a specific officer
func (s *service) GetOfficerDashboard(ctx context.Context, officerID string, period string) (*OfficerDashboardResponse, error) {
	s.logger.Info("Getting officer dashboard", zap.String("officerID", officerID), zap.String("period", period))

	// Verify officer exists
	_, err := s.userRepo.GetByID(ctx, officerID)
	if err != nil {
		return nil, err
	}

	uid, _ := uuid.Parse(officerID)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	// Get controles statistics
	controlsTotal := s.countControles(ctx, uid, nil)
	controlsToday := s.countControles(ctx, uid, &todayStart)
	controlsWeek := s.countControles(ctx, uid, &weekStart)
	controlsMonth := s.countControles(ctx, uid, &monthStart)
	controlsYear := s.countControles(ctx, uid, &yearStart)

	// Get inspections statistics
	inspectionsTotal := s.countInspections(ctx, uid, nil)
	inspectionsToday := s.countInspections(ctx, uid, &todayStart)
	inspectionsWeek := s.countInspections(ctx, uid, &weekStart)
	inspectionsMonth := s.countInspections(ctx, uid, &monthStart)
	inspectionsYear := s.countInspections(ctx, uid, &yearStart)

	// Get revenue statistics from PVs
	revenueTotal := s.getRevenue(ctx, uid, nil)
	revenueToday := s.getRevenue(ctx, uid, &todayStart)
	revenueWeek := s.getRevenue(ctx, uid, &weekStart)
	revenueMonth := s.getRevenue(ctx, uid, &monthStart)
	revenueYear := s.getRevenue(ctx, uid, &yearStart)

	// Get compliance statistics
	conformeCount := s.countInspectionsConforme(ctx, uid)
	conformanceRate := 0.0
	if inspectionsTotal > 0 {
		conformanceRate = float64(conformeCount) / float64(inspectionsTotal) * 100
	}

	// Get top infractions
	infractionStats, err := s.infractionRepo.GetStatistics(ctx, &repository.InfractionStatsFilters{AgentID: &officerID})
	var topInfractions []InfractionRank
	if err == nil && infractionStats != nil {
		for i, inf := range infractionStats.TopInfractions {
			if i >= 5 {
				break
			}
			topInfractions = append(topInfractions, InfractionRank{
				Name:  inf.TypeLibelle,
				Code:  inf.TypeCode,
				Count: inf.Count,
				Rank:  i + 1,
			})
		}
	}

	// Build activity chart (last 7 days)
	activityChart := s.buildActivityChart(ctx, uid, 7)

	return &OfficerDashboardResponse{
		Controls: PeriodStats{
			Total: controlsTotal,
			Today: controlsToday,
			Week:  controlsWeek,
			Month: controlsMonth,
			Year:  controlsYear,
		},
		Inspections: PeriodStats{
			Total: inspectionsTotal,
			Today: inspectionsToday,
			Week:  inspectionsWeek,
			Month: inspectionsMonth,
			Year:  inspectionsYear,
		},
		Revenues: RevenueStats{
			Total: revenueTotal,
			Today: revenueToday,
			Week:  revenueWeek,
			Month: revenueMonth,
			Year:  revenueYear,
		},
		Compliance: ComplianceStats{
			VehiclesConformed: conformeCount,
			VehiclesInspected: inspectionsTotal,
			ConformanceRate:   conformanceRate,
		},
		Infractions:   topInfractions,
		ActivityChart: activityChart,
		Period:        period,
	}, nil
}

// GetOfficerStatistics returns simple statistics for a specific officer
func (s *service) GetOfficerStatistics(ctx context.Context, officerID string) (*OfficerStatisticsResponse, error) {
	s.logger.Info("Getting officer statistics", zap.String("officerID", officerID))

	// Verify officer exists
	_, err := s.userRepo.GetByID(ctx, officerID)
	if err != nil {
		return nil, err
	}

	uid, _ := uuid.Parse(officerID)

	// Get total PVs created by this officer (via controle or inspection)
	totalPVs, _ := s.client.ProcesVerbal.Query().
		Where(
			procesverbal.Or(
				procesverbal.HasControleWith(controle.HasAgentWith(user.ID(uid))),
				procesverbal.HasInspectionWith(inspection.HasInspecteurWith(user.ID(uid))),
			),
		).
		Count(ctx)

	// Get active PVs (not paid)
	activePVs, _ := s.client.ProcesVerbal.Query().
		Where(
			procesverbal.Or(
				procesverbal.HasControleWith(controle.HasAgentWith(user.ID(uid))),
				procesverbal.HasInspectionWith(inspection.HasInspecteurWith(user.ID(uid))),
			),
			procesverbal.StatutEQ("EN_ATTENTE"),
		).
		Count(ctx)

	// Get paid PVs count
	paidPVs, _ := s.client.ProcesVerbal.Query().
		Where(
			procesverbal.Or(
				procesverbal.HasControleWith(controle.HasAgentWith(user.ID(uid))),
				procesverbal.HasInspectionWith(inspection.HasInspecteurWith(user.ID(uid))),
			),
			procesverbal.StatutEQ("PAYE"),
		).
		Count(ctx)

	// Get total fine amount
	pvs, _ := s.client.ProcesVerbal.Query().
		Where(
			procesverbal.Or(
				procesverbal.HasControleWith(controle.HasAgentWith(user.ID(uid))),
				procesverbal.HasInspectionWith(inspection.HasInspecteurWith(user.ID(uid))),
			),
		).
		All(ctx)

	var totalAmount float64
	for _, pv := range pvs {
		totalAmount += pv.MontantTotal
	}

	return &OfficerStatisticsResponse{
		TotalIssuedTickets: totalPVs,
		ActiveTickets:      activePVs,
		TotalPayments:      paidPVs,
		TotalFineAmount:    totalAmount,
	}, nil
}

// Helper: count controles for officer after a date
func (s *service) countControles(ctx context.Context, officerID uuid.UUID, after *time.Time) int {
	query := s.client.Controle.Query().
		Where(controle.HasAgentWith(user.ID(officerID)))

	if after != nil {
		query = query.Where(controle.DateControleGTE(*after))
	}

	count, _ := query.Count(ctx)
	return count
}

// Helper: count inspections for officer after a date
func (s *service) countInspections(ctx context.Context, officerID uuid.UUID, after *time.Time) int {
	query := s.client.Inspection.Query().
		Where(inspection.HasInspecteurWith(user.ID(officerID)))

	if after != nil {
		query = query.Where(inspection.DateInspectionGTE(*after))
	}

	count, _ := query.Count(ctx)
	return count
}

// Helper: count conforme inspections
func (s *service) countInspectionsConforme(ctx context.Context, officerID uuid.UUID) int {
	count, _ := s.client.Inspection.Query().
		Where(
			inspection.HasInspecteurWith(user.ID(officerID)),
			inspection.StatutEQ(inspection.StatutCONFORME),
		).
		Count(ctx)
	return count
}

// Helper: get revenue from PVs for officer after a date
func (s *service) getRevenue(ctx context.Context, officerID uuid.UUID, after *time.Time) float64 {
	query := s.client.ProcesVerbal.Query().
		Where(
			procesverbal.Or(
				procesverbal.HasControleWith(controle.HasAgentWith(user.ID(officerID))),
				procesverbal.HasInspectionWith(inspection.HasInspecteurWith(user.ID(officerID))),
			),
		)

	if after != nil {
		query = query.Where(procesverbal.DateEmissionGTE(*after))
	}

	pvs, _ := query.All(ctx)
	var total float64
	for _, pv := range pvs {
		total += pv.MontantTotal
	}
	return total
}

// Helper: build activity chart
func (s *service) buildActivityChart(ctx context.Context, officerID uuid.UUID, days int) []int {
	chart := make([]int, days)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		dayStart := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, now.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)

		count, _ := s.client.Controle.Query().
			Where(
				controle.HasAgentWith(user.ID(officerID)),
				controle.DateControleGTE(dayStart),
				controle.DateControleLT(dayEnd),
			).
			Count(ctx)

		chart[days-1-i] = count
	}

	return chart
}
