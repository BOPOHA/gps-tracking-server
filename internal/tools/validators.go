package tools

func IsValidCoordinates(latFloat float64, lonFloat float64) bool {
	if lonFloat == 0 || lonFloat < -180 || lonFloat > 180 || latFloat == 0 || latFloat < -90 || latFloat > 90 {
		return false
	}
	return true
}

func IsValidRecord(satellites byte) bool {
	return satellites > 3
}
