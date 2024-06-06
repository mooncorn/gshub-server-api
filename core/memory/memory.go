package memory

func CalculateServerMemory(maxAvailableMemoryMB int) int {
	return maxAvailableMemoryMB - 512 // 512mb allocated for the system and api
}

func CheckMinimumMemory(serverMemoryMB int, minimumServerMemoryMB int) bool {
	return minimumServerMemoryMB <= serverMemoryMB
}
