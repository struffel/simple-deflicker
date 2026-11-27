package deflicker

import "math"

// Rec. 709 luma weights.
const (
	wR = 0.2126
	wG = 0.7152
	wB = 0.0722
)

// histogramComparison describes how histogram B differs from histogram A.
type histogramComparison struct {
	Distance        float64    // luma-weighted unsigned EMD, normalized to [0, 1]
	Shift           float64    // luma-weighted signed shift, in [-1, 1]; >0 = B lighter, <0 = B darker
	Consistency     float64    // |Shift| / Distance in [0, 1]; ~1 = pure brightness change, ~0 = contrast/shape change
	PerChannelShift [3]float64 // signed shift for R, G, B
	PerChannelDist  [3]float64 // unsigned EMD for R, G, B
}

// channelEMD returns the unsigned and signed 1D earth mover's distance between
// two histograms, both normalized by the bin count so the result is in [0, 1].
// The signed value equals mean(B) - mean(A), scaled to [-1, 1].
// An empty histogram yields (0, 0).
func channelEMD(a, b *histogram) (earthMoversDistance float64, shift float64) {
	var totalA, totalB float64
	for i := range a {
		totalA += float64(a[i])
		totalB += float64(b[i])
	}
	if totalA == 0 || totalB == 0 {
		return 0, 0
	}

	var cumA, cumB float64
	for i := range a {
		cumA += float64(a[i]) / totalA
		cumB += float64(b[i]) / totalB
		d := cumA - cumB // >0: A has more dark pixels so far, so B is lighter
		earthMoversDistance += math.Abs(d)
		shift += d
	}
	n := float64(len(a) - 1) // 255
	return earthMoversDistance / n, shift / n
}

func compareHistograms(a, b rgbHistogram) histogramComparison {
	var res histogramComparison

	res.PerChannelDist[0], res.PerChannelShift[0] = channelEMD(&a.R, &b.R)
	res.PerChannelDist[1], res.PerChannelShift[1] = channelEMD(&a.G, &b.G)
	res.PerChannelDist[2], res.PerChannelShift[2] = channelEMD(&a.B, &b.B)

	res.Distance = wR*res.PerChannelDist[0] + wG*res.PerChannelDist[1] + wB*res.PerChannelDist[2]
	res.Shift = wR*res.PerChannelShift[0] + wG*res.PerChannelShift[1] + wB*res.PerChannelShift[2]

	if res.Distance > 1e-12 {
		res.Consistency = math.Abs(res.Shift) / res.Distance
	} else {
		res.Consistency = 1 // identical histograms: nothing inconsistent
	}
	return res
}
