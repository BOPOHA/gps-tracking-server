package tools

func IsValidCoordinates(lat_float float64, lon_float float64) bool {
	if lon_float == 0 || lon_float < -180 || lon_float > 180 || lat_float == 0 || lat_float < -90 || lat_float > 90 {
		return false
	}
	return true
}

func IsValidRecord(satellites byte) bool {
	return (satellites > 3)
}

