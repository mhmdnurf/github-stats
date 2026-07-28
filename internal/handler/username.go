package handler

func validGitHubUsername(username string) bool {
	if len(username) == 0 || len(username) > 39 {
		return false
	}

	for index := 0; index < len(username); index++ {
		character := username[index]

		isLetter :=
			character >= 'a' && character <= 'z' ||
				character >= 'A' && character <= 'Z'
		isNumber :=
			character >= '0' && character <= '9'

		if isLetter || isNumber {
			continue
		}

		if character != '-' {
			return false
		}

		if index == 0 || index == len(username)-1 {
			return false
		}

		if username[index-1] == '-' {
			return false
		}
	}

	return true
}
