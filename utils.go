package main

func getDayFromSlotNo(slotNo int) string {
	if slotNo >= 1 && slotNo <= 6 {
		return "Monday"
	} else if slotNo >= 7 && slotNo <= 12 {
		return "Tuesday"
	} else if slotNo >= 13 && slotNo <= 18 {
		return "Wednesday"
	} else if slotNo >= 19 && slotNo <= 24 {
		return "Thursday"
	} else if slotNo >= 25 && slotNo <= 30 {
		return "Friday"
	} else if slotNo >= 31 && slotNo <= 38 {
		return "Saturday"
	} else if slotNo >= 39 && slotNo <= 46 {
		return "Sunday"
	}
	return ""
}

func getSlotStartHour(slotno int) int {
	if slotno < 30 {
		slotno = slotno % 6
		switch slotno {
		case 1:
			return 7
		case 2:
			return 15
		case 3:
			return 17
		case 4:
			return 19
		case 5:
			return 21
		case 0:
			return 23
		}
	} else {
		slotno -= 30
		slotno = slotno % 8
		switch slotno {
		case 1:
			return 7
		case 2:
			return 9
		case 3:
			return 11
		case 4:
			return 13
		case 5:
			return 15
		case 6:
			return 17
		case 7:
			return 19
		case 0:
			return 21
		}
	}
	return -1
}
