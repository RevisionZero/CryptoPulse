package engine

import (
	"log"
	"main/pkg/models"
	"math"
)

type PCCState struct {
	meanX, meanY  float64
	sXX, sYY, sXY float64
	count         float64 // Tracks if the window is full (e.g., 600)
}

func (s *PCCState) Update(xNew, yNew, xOld, yOld float64, windowSize float64) float64 {
	// Step 1: INVERSE (Subtract the old point)
	// We only subtract if we have already reached the full window size
	if s.count >= windowSize {
		// Shift means "backwards" to remove the old point's influence
		muX_prev := (windowSize*s.meanX - xOld) / (windowSize - 1)
		muY_prev := (windowSize*s.meanY - yOld) / (windowSize - 1)

		// Subtract from Sum of Squares and Cross-Products
		s.sXX -= (xOld - s.meanX) * (xOld - muX_prev)
		s.sYY -= (yOld - s.meanY) * (yOld - muY_prev)
		s.sXY -= (xOld - s.meanX) * (yOld - muY_prev)

		s.meanX, s.meanY = muX_prev, muY_prev
	} else {
		s.count++
	}

	// Step 2: FORWARD (Add the new point - Standard Welford)
	dx := xNew - s.meanX
	dy := yNew - s.meanY
	s.meanX += dx / s.count
	s.meanY += dy / s.count

	s.sXX += dx * (xNew - s.meanX)
	s.sYY += dy * (yNew - s.meanY)
	s.sXY += dx * (yNew - s.meanY)

	// Step 3: Calculate Result
	denom := math.Sqrt(s.sXX * s.sYY)
	if denom == 0 {
		return 0
	}
	return s.sXY / denom
}

// func PCCMatrixCalculator(sampledDataChan <-chan map[string][]float64, symbols []string) {
// 	pccMatrix := make(map[string]map[string]float64, len(symbols))
// 	for _, symbolX := range symbols {
// 		pccMatrix[symbolX] = make(map[string]float64, len(symbols))
// 		for _, symbolY := range symbols {
// 			if symbolX == symbolY {
// 				pccMatrix[symbolX][symbolY] = 1.0
// 			}
// 		}
// 	}
// 	for {
// 		sampledData := <-sampledDataChan
// 		CalculatePCCMatrix(sampledData, symbols, pccMatrix)
// 		_ = pccMatrix
// 		// log.Printf("PCC Matrix: %+v\n", pccMatrix)

// 	}
// }

func PCCMatrixCalculator(mutationChan <-chan models.PriceMutation, symbols []string) {
	// Store the running math for every pair
	pairStates := make(map[string]*PCCState)
	pccMatrix := make(map[string]map[string]float64, len(symbols))
	for _, symbolX := range symbols {
		pccMatrix[symbolX] = make(map[string]float64, len(symbols))
		for _, symbolY := range symbols {
			if symbolX == symbolY {
				pccMatrix[symbolX][symbolY] = 1.0
			}
		}
	}

	for mutation := range mutationChan {
		for i, sX := range symbols {
			for j := i + 1; j < len(symbols); j++ {
				sY := symbols[j]
				pairKey := sX + ":" + sY

				if _, exists := pairStates[pairKey]; !exists {
					pairStates[pairKey] = &PCCState{}
				}

				// Update only what changed!
				val := pairStates[pairKey].Update(
					mutation.NewPrices[sX], mutation.NewPrices[sY],
					mutation.OldPrices[sX], mutation.OldPrices[sY],
					600.0,
				)

				pccMatrix[sX][sY] = val
				pccMatrix[sY][sX] = val
			}
		}

		log.Printf("PCC Matrix: %+v\n", pccMatrix)
	}
}

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

func CalculatePCCMatrix(sampledData map[string][]float64, symbols []string, pccMatrix map[string]map[string]float64) {

	for _, symbolX := range symbols {
		for _, symbolY := range symbols {
			if symbolX == symbolY {
				pccMatrix[symbolX][symbolY] = 1.0
			} else {
				pccValue := PCC(sampledData[symbolX], sampledData[symbolY], len(sampledData[symbolX]))
				pccMatrix[symbolX][symbolY] = pccValue
			}
		}
	}

}
