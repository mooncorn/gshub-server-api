package memory

func CalculateServerMemory(maxAvailableMemoryMB int) int {
	return maxAvailableMemoryMB - 1024 // 1GB allocated for the system and api
}

func CheckMinimumMemory(serverMemoryMB int, minimumServerMemoryMB int) bool {
	return minimumServerMemoryMB <= serverMemoryMB
}
