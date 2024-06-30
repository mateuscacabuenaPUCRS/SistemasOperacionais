package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
)

const (
	// Print Intermediate Steps
	DEBUG = false

	DEFAULT_V_MEM_SIZE = 16 // 40
	DEFAULT_F_MEM_SIZE = 15 // 37
	DEFAULT_PAGE_SIZE = 12 // 25
	DEFAULT_SEED = 42
	DEFAULT_V_ADDRS_COUNT = 10

	// Should match type "number" (uintMAX_NUMBER)
	MAX_NUMBER = 64

	// Console Editing
	BOLD = "\033[1m"
	RED = "\033[31m"
	GREEN = "\033[32m"
	YELLOW = "\033[33m"
	BLUE = "\033[34m"
	MAGENTA = "\033[35m"
	CYAN = "\033[36m"
	RESET = "\033[0m"
)

var (
	// Virtual Memory Size
	V_MEM_SIZE number
	// Physical Memory Size
	F_MEM_SIZE number
	// Page Size
	PAGE_SIZE number

	// Physical Memory
	F_MEM []number
	// Page Table
	PAGE_TABLE []int

	// Random Seed
	RAND_SEED number
	// Virtual Addresses
	V_ADDRS []number
)

type number = uint64

// Returns the argument at the given index or the default value if it was not provided
// Exits the program if the argument is not a non-negative integer
func getArg(index int, defaul number) number {
	if len(os.Args) > index {
		value, err := strconv.Atoi(os.Args[index])
		if err != nil || value < 0 || value > MAX_NUMBER {
			fmt.Println(colorize(RED, fmt.Sprintf(
				"Invalid argument %d: %s. Must be a non-negative integer up to %d", index, os.Args[index], MAX_NUMBER)))
			os.Exit(1)
		}
		return number(value)
	}
	return defaul
}

// Calculates 2 to the power of the given number and returns it
func pow2(n number) number {
	power := math.Pow(2, float64(n))
	return number(power)
}

// Converts bits to bytes
func bitsToBytes(bits number) number {
	return bits / 8
}

// Returns the given number of memory formatted as Kilo, Mega, Giga or Tera {unit}
// @param unit: bits or bytes
func formatMemory(memory number, unit string) string {
	var ending string
	if unit == "bits" {
		ending = "b"
	} else if unit == "bytes" {
		ending = "B"
	} else {
		os.Exit(1)
	}

	if memory < pow2(10) {
		return fmt.Sprintf("%d %s", int(memory), unit)
	}
	if memory < pow2(20) {
		return fmt.Sprintf("%d K%s", memory/1024, ending)
	}
	if memory < pow2(30) {
		return fmt.Sprintf("%d M%s", memory/1024/1024, ending)
	}
	if float64(memory) < math.Pow(2, 40) {
		return fmt.Sprintf("%d G%s", memory/1024/1024/1024, ending)
	}
	return fmt.Sprintf("%d T%s", memory/1024/1024/1024/1024, ending)
}

// Returns a string formatted to be printed with color
func colorize(color string, text string) string {
	return fmt.Sprintf("%s%s%s", color, text, RESET)
}

// Handles the arguments passed to the program
// Exits the program if the arguments are invalid
func handleArgs() {
	fmt.Println()

	if len(os.Args) < 6 {
		fmt.Println("\nUseful arguments missing")
		fmt.Printf("Usage: go run SGMP.go %s <V = Virtual Memory Size> %s <F = Physical Memory Size> %s <P = Page Size> %s <Optional: S = Random Seed> %s <Optional: A = The Amount of Virtual Addresses to Generate if One, the Virtual Addresses if More>\n", CYAN, MAGENTA, GREEN, BLUE, YELLOW)
		fmt.Println(RESET, BOLD)
		fmt.Printf("All sizes are in 2^n (bits). i.e. 2^%d = %s = %s\n",
			10, formatMemory(pow2(10), "bits"), formatMemory(bitsToBytes(pow2(10)), "bytes"))
		fmt.Printf("Example: go run SGMP.go %s %s %s %s %s\n",
			colorize(CYAN, fmt.Sprint(DEFAULT_V_MEM_SIZE)),
			colorize(MAGENTA, fmt.Sprint(DEFAULT_F_MEM_SIZE)),
			colorize(GREEN, fmt.Sprint(DEFAULT_PAGE_SIZE)),
			colorize(BLUE, fmt.Sprint(DEFAULT_SEED)),
			colorize(YELLOW, fmt.Sprint(DEFAULT_V_ADDRS_COUNT)),
		)
		fmt.Print(BOLD)
		fmt.Println("Using default values to fill non-provided arguments")
	}
	
	V_MEM_SIZE = getArg(1, DEFAULT_V_MEM_SIZE)
	F_MEM_SIZE = getArg(2, DEFAULT_F_MEM_SIZE)
	PAGE_SIZE = getArg(3, DEFAULT_PAGE_SIZE)
	RAND_SEED = getArg(4, DEFAULT_SEED)
	randomizer := rand.New(rand.NewSource(int64(RAND_SEED)))

	fmt.Println(colorize(CYAN,
		fmt.Sprintf("\tVirtual Memory Size:\t2^%d = %s = %s", V_MEM_SIZE,
			formatMemory(pow2(V_MEM_SIZE), "bits"), formatMemory(bitsToBytes(pow2(V_MEM_SIZE)), "bytes")),
	))

	fmt.Println(colorize(MAGENTA,
		fmt.Sprintf("\tPhysical Memory Size:\t2^%d = %s = %s", F_MEM_SIZE,
			formatMemory(pow2(F_MEM_SIZE), "bits"), formatMemory(bitsToBytes(pow2(F_MEM_SIZE)), "bytes")),
	))

	fmt.Println(colorize(GREEN,
		fmt.Sprintf("\tPage Size:\t\t2^%d = %s = %s", PAGE_SIZE,
			formatMemory(pow2(PAGE_SIZE), "bits"), formatMemory(bitsToBytes(pow2(PAGE_SIZE)), "bytes")),
	))

	fmt.Println(RESET)
	upper_v_addr_limit := number(pow2(V_MEM_SIZE) - 1)
	if len(os.Args) > 6 {
		V_ADDRS = make([]number, len(os.Args) - 5)
		for i := 5; i < len(os.Args); i++ {
			value, err := strconv.Atoi(os.Args[i])
			if err != nil || value < 0 || number(value) > upper_v_addr_limit {
				fmt.Println(colorize(RED, fmt.Sprintf(
					"Invalid argument %d: %s. Must be a non-negative integer up to %d (2^V - 1)", i, os.Args[i], upper_v_addr_limit)))
				os.Exit(1)
			}
			V_ADDRS[i-5] = number(value)
		}
		fmt.Printf("Using the following arguments as virtual addresses: %d\n", V_ADDRS)
	} else {
		V_ADDRS_COUNT := getArg(5, DEFAULT_V_ADDRS_COUNT)
		fmt.Printf("Using %s as random seed to generate %s virtual addresses\n",
			colorize(BLUE, fmt.Sprint(RAND_SEED)), colorize(YELLOW, fmt.Sprint(V_ADDRS_COUNT)))
		V_ADDRS = make([]number, V_ADDRS_COUNT)
		for i := 0; i < len(V_ADDRS); i++ {
			V_ADDRS[i] = number(randomizer.Uint32() % uint32(upper_v_addr_limit))
		}
		fmt.Printf("Generated virtual addresses: %d\n", V_ADDRS)
	}
}

func setupTables() {
	fmt.Println()

	pages_count := pow2(V_MEM_SIZE - PAGE_SIZE)
	frames_count := pow2(F_MEM_SIZE - PAGE_SIZE)
	fmt.Printf("Page Table has %d pages of %s (%s) each\n", pages_count,
		formatMemory(pow2(PAGE_SIZE), "bits"), formatMemory(bitsToBytes(pow2(PAGE_SIZE)), "bytes"))
	fmt.Printf("Physical Memory has %d frames of %s (%s) each\n", frames_count,
		formatMemory(pow2(PAGE_SIZE), "bits"), formatMemory(bitsToBytes(pow2(PAGE_SIZE)), "bytes"))
	PAGE_TABLE = make([]int, pages_count)
	F_MEM = make([]number, frames_count)

	for i := 0; i < len(PAGE_TABLE); i++ {
		PAGE_TABLE[i] = -1
	}

	fmt.Printf("Physical Memory and Page Table initialized with %ds and %ds, respectively\n", 0, -1)
}

// // @Deprecated
// func splitPageShift(virtual_address number) (page number, shift number) {
// 	page_length := int(math.Log2(float64(len(F_MEM))))
// 	virtual_address_str := strconv.Itoa(int(virtual_address))
// 	page_str := virtual_address_str[:page_length]
// 	shift_str := virtual_address_str[page_length:]
// 	page_i, _ := strconv.Atoi(page_str)
// 	shift_i, _ := strconv.Atoi(shift_str)
// 	return number(page_i), number(shift_i)
// }

func mapVirtualToPhysicalAddress(virtualAddress number) (physicalAddress number) {
	pageSizeBits := pow2(PAGE_SIZE)
	pageIndex := virtualAddress / pageSizeBits
	shift := virtualAddress % pageSizeBits
	frameIndex := PAGE_TABLE[pageIndex]
	if frameIndex == -1 {
		physicalMemoryFull := true
		for i := 0; i < len(F_MEM); i++ {
			if F_MEM[i] == 0 {
				PAGE_TABLE[pageIndex] = i
				frameIndex = i
				physicalMemoryFull = false
				break
			}
		}
		if physicalMemoryFull {
			fmt.Println(colorize(RED, fmt.Sprintln(
				"Physical Memory is Full, Program Stopping...")))
			os.Exit(1)
		}
	}
	F_MEM[frameIndex] = 1 // virtualAddress
	frameStartingAddress := number(frameIndex) * pageSizeBits
	physicalAddress = frameStartingAddress + shift
	return physicalAddress
}

func main() {
	fmt.Println(RESET, BOLD)
	fmt.Println("=====", "Sistema Gerência de Memória Paginada", "=====")

	handleArgs()
	setupTables()

	fmt.Println(RESET, BOLD)
	fmt.Println(colorize(YELLOW, fmt.Sprint(
		"====================", " Starting Conversions ", "====================")))

	var physicalAddress number
	physicalAddresses := make([]number, len(V_ADDRS))
	for i := 0; i < len(V_ADDRS); i++ {
		physicalAddress = mapVirtualToPhysicalAddress(V_ADDRS[i])
		physicalAddresses[i] = physicalAddress
		if DEBUG {
			fmt.Printf("\nVirtual Address %d: %d\n", i, V_ADDRS[i])
			fmt.Printf("Physical Address %d: %d\n", i, physicalAddress)
			fmt.Println("Page Table: ", PAGE_TABLE)
			fmt.Println("Physical Memory: ", F_MEM)
		}
	}

	fmt.Println(RESET, BOLD)
	fmt.Println(colorize(YELLOW, fmt.Sprint(
		"====================", " OUTPUT ", "====================")))

	fmt.Println(colorize(CYAN, fmt.Sprint(
		"Virtual Addresses:\t", V_ADDRS)))
	fmt.Println(colorize(MAGENTA, fmt.Sprint(
		"Physical Addresses:\t", physicalAddresses)))
	fmt.Println(colorize(GREEN, fmt.Sprint(
		"Page Table:\t\t", PAGE_TABLE)))
	fmt.Println(colorize(MAGENTA, fmt.Sprint(
		"Physical Memory:\t", F_MEM)))

	fmt.Println(RESET)
}
