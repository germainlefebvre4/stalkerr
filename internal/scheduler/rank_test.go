package scheduler

import (
	"testing"

	"github.com/glefebvre/stalkeer/internal/models"
)

func TestCandidateRankFrenchVariantBreaksTieWithinLanguageAndResolution(t *testing.T) {
	nonVFQ := &models.ProcessedLine{Language: strPtr("VF"), Resolution: strPtr("720p")}
	vfq := &models.ProcessedLine{Language: strPtr("VF"), Resolution: strPtr("720p"), FrenchVariant: strPtr("VFQ")}

	if candidateRank(nonVFQ) >= candidateRank(vfq) {
		t.Errorf("expected non-VFQ candidate rank (%d) to be lower than VFQ candidate rank (%d)",
			candidateRank(nonVFQ), candidateRank(vfq))
	}
}

func TestCandidateRankFrenchVariantNoEffectWhenLanguageDiffers(t *testing.T) {
	vfNonVFQ := &models.ProcessedLine{Language: strPtr("VF"), Resolution: strPtr("720p")}
	multiVFQ := &models.ProcessedLine{Language: strPtr("MULTI"), Resolution: strPtr("720p"), FrenchVariant: strPtr("VFQ")}

	// Language still dominates: VF ranks better than MULTI regardless of VFQ.
	if candidateRank(vfNonVFQ) >= candidateRank(multiVFQ) {
		t.Errorf("expected VF candidate rank (%d) to be lower than MULTI/VFQ candidate rank (%d)",
			candidateRank(vfNonVFQ), candidateRank(multiVFQ))
	}
}

func TestCandidateRankFrenchVariantNoEffectWhenResolutionDiffers(t *testing.T) {
	res720NonVFQ := &models.ProcessedLine{Language: strPtr("VF"), Resolution: strPtr("720p")}
	res4KVFQ := &models.ProcessedLine{Language: strPtr("VF"), Resolution: strPtr("4K"), FrenchVariant: strPtr("VFQ")}

	// Resolution still dominates: 720p ranks better than 4K regardless of VFQ.
	if candidateRank(res720NonVFQ) >= candidateRank(res4KVFQ) {
		t.Errorf("expected 720p candidate rank (%d) to be lower than 4K/VFQ candidate rank (%d)",
			candidateRank(res720NonVFQ), candidateRank(res4KVFQ))
	}
}
