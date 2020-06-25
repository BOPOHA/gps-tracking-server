package tools


// HELPERS

func PadLeft(str, pad string, lenght int) string {

	if len(str) >= lenght {
		return str
	}

	for {
		str = pad + str
		if len(str) >= lenght {
			return str
		}
	}
}

