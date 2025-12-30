package engine

import (
	"log"
	"math"
)

func PCCMatrixCalculator(sampledDataChan <-chan map[string][]float64, symbols []string, matrixChan chan<- map[string]map[string]float64) {
	pccMatrix := make(map[string]map[string]float64, len(symbols))
	for _, symbolX := range symbols {
		pccMatrix[symbolX] = make(map[string]float64, len(symbols))
		for _, symbolY := range symbols {
			if symbolX == symbolY {
				pccMatrix[symbolX][symbolY] = 1.0
			}
		}
	}
	for {
		sampledData := <-sampledDataChan
		CalculatePCCMatrix(sampledData, symbols, pccMatrix)
		log.Println("PCC Matrix calculated,%s", pccMatrix)
		matrixChan <- pccMatrix

	}
}

func PCC(x []float64, y []float64, sampleSize int) float64 {

	sumOfProducts := 0.0

	diffOfSquaresX := 0.0
	diffOfSquaresY := 0.0
	meanX := 0.0
	meanY := 0.0

	for i := range x {
		oldMeanX := meanX
		oldMeanY := meanY

		meanX = meanX + (x[i]-meanX)/float64(i+1)
		meanY = meanY + (y[i]-meanY)/float64(i+1)

		sumOfProducts += (x[i] - oldMeanX) * (y[i] - meanY)

		diffOfSquaresX = diffOfSquaresX + (x[i]-meanX)*(x[i]-oldMeanX)
		diffOfSquaresY = diffOfSquaresY + (y[i]-meanY)*(y[i]-oldMeanY)
	}

	result := sumOfProducts / math.Sqrt(diffOfSquaresX*diffOfSquaresY)
	if math.IsNaN(result) {
		return 0.0
	}
	return result
}

func CalculatePCCMatrix(sampledData map[string][]float64, symbols []string, pccMatrix map[string]map[string]float64) {

	for _, symbolX := range symbols {
		for _, symbolY := range symbols {
			if symbolX == symbolY {
				pccMatrix[symbolX][symbolY] = 1.0
			} else {
				dataX, okX := sampledData[symbolX]
				dataY, okY := sampledData[symbolY]

				// Check if both data slices exist and have data
				if !okX || !okY || len(dataX) == 0 || len(dataY) == 0 {
					continue
				}

				pccValue := PCC(dataX, dataY, len(dataX))
				pccMatrix[symbolX][symbolY] = pccValue
			}
		}
	}

}
