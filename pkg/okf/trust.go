package okf

import (
	"strings"
	"time"
)

type TrustTier string

const (
	TierStale            TrustTier = "Stale"
	TierHumanReviewed    TrustTier = "Human-Reviewed"
	TierMachineConfirmed TrustTier = "Machine-Confirmed"
	TierUnverified       TrustTier = "Unverified"
)

func EvaluateTrustTier(fm *Frontmatter, now time.Time) TrustTier {
	if fm.StaleAfter != "" {
		staleTime, err := time.Parse("2006-01-02", fm.StaleAfter)
		if err == nil && now.After(staleTime) {
			return TierStale
		}
	}

	for _, v := range fm.Verified {
		if strings.HasPrefix(string(v.Actor), "human:") {
			return TierHumanReviewed
		}
	}

	for _, v := range fm.Verified {
		if strings.HasPrefix(string(v.Actor), "process:") || strings.Contains(string(v.Actor), "/") {
			return TierMachineConfirmed
		}
	}

	return TierUnverified
}
