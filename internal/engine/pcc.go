package engine

import "math"

func PCC(x []float64, y []float64, sampleSize int) float64 {

	sumOfProducts := 0.0

	diffOfSquaresX := 0.0
	diffOfSquaresY := 0.0
	meanX := 0.0
	meanY := 0.0

	for i, _ := range x {
		oldMeanX := meanX
		oldMeanY := meanY

		meanX = meanX + (x[i]-meanX)/float64(i+1)
		meanY = meanY + (y[i]-meanY)/float64(i+1)

		sumOfProducts += (x[i] - oldMeanX) * (y[i] - meanY)

		diffOfSquaresX = diffOfSquaresX + (x[i]-meanX)*(x[i]-oldMeanX)
		diffOfSquaresY = diffOfSquaresY + (y[i]-meanY)*(y[i]-oldMeanY)
	}
	return sumOfProducts / math.Sqrt(diffOfSquaresX*diffOfSquaresY)
}

// func CalculatePCCMatrix(sampledData map[string][]float64, symbols []string) map[string]map[string]float64 {
// 	pccMatrix := make(map[string]map[string]float64)

// 	for _, symbolX := range symbols {
// 		pccMatrix[symbolX] = make(map[string]float64)
// 		for _, symbolY := range symbols {
// 			if symbolX == symbolY {
// 				pccMatrix[symbolX][symbolY] = 1.0
// 			} else {
// 				pccValue := PCC(sampledData[symbolX], sampledData[symbolY], len(sampledData[symbolX]))
// 				pccMatrix[symbolX][symbolY] = pccValue
// 			}
// 		}
// 	}
// 	return pccMatrix
// }
