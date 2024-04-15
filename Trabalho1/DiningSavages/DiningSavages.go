package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	DEFAULT_COOKS_COUNT = 1
	DEFAULT_SAVAGES_COUNT = 5
	DEFAULT_THREAD_COUNT = DEFAULT_COOKS_COUNT + DEFAULT_SAVAGES_COUNT
	DEFAULT_SERVINGS_COUNT = 5

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
	// Savages Count
	N uint
	// Max Servings Count
	maxServingsCount uint
	// Variable Servings Count
	M uint
	// Cooks Count
	L uint
	// Threads Count
	threadCount uint

	mutex *Mutex
	emptyPot = NewSemaphore(0)
	fullPot = NewSemaphore(0)
)

// Returns the argument at the given index or the default value if it was not provided
// Exits the program if the argument is not a non-negative integer
func getArg(index int, defaul uint) uint {
	if len(os.Args) > index {
		value, err := strconv.Atoi(os.Args[index])
		if err != nil || value < 0 {
			fmt.Printf(
				"%s Invalid argument %d: %d. Must be a non-negative integer. %s\n",
				RED, index, value, RESET,
			)
			os.Exit(1)
		}
		return uint(value)
	}
	return defaul
}

// Returns a random integer between 0 and max
func GetRandom(max uint) int {
	return rand.Intn(int(max) + 1)
}

// Waits for a random time in milliseconds with 10x a custom multiplier
func WaitRandom(multiplier uint) {
	time.Sleep(time.Duration(GetRandom(10 * multiplier)) * time.Millisecond)
}

func Cook(id uint) {
	for {
		emptyPot.Wait()
		putServingsInPot(id)
		M = maxServingsCount
		fullPot.Signal()
		WaitRandom(DEFAULT_COOKS_COUNT)
	}
}

func putServingsInPot(id uint) {
	PrintCook(fmt.Sprint("Cook ", id, " is putting servings in pot...", "\n"))
}

func Savage(id uint) {
	for {
		WaitRandom(DEFAULT_SAVAGES_COUNT)
		mutex.Lock(id)
			if M == 0 {
				wakeUpCook(id)
				emptyPot.Signal()
				fullPot.Wait()
			}
			M--
			getServingFromPot(id)
		mutex.Unlock(id)
	}
}

func wakeUpCook(id uint) {
	PrintAlert(fmt.Sprint("Savage ", id, "- Pot is empty, I'll wake up the cook...", "\n"))
}

func getServingFromPot(id uint) {
	PrintSavage(fmt.Sprint("Savage ", id, " is serving..."))
}

func main() {
	fmt.Println(RESET, BOLD)
	fmt.Println("=====", "Dining Savages", "=====")

	if len(os.Args) < 4 {
		fmt.Println("\nUseful arguments missing")
		fmt.Printf("Usage: go run DiningSavages.go %s <N = number of savages> %s <M = number of servings> %s <Optional: L = number of cooks>\n", CYAN, MAGENTA, GREEN)
		fmt.Print(RESET, BOLD)
		fmt.Printf("Example: go run DiningSavages.go %s %d %s %d %s %d\n", CYAN, DEFAULT_SAVAGES_COUNT, MAGENTA, DEFAULT_SERVINGS_COUNT, GREEN, DEFAULT_COOKS_COUNT)
		fmt.Print(RESET, BOLD)
		fmt.Println("Using default values to fill non-provided arguments")
	}
	N = getArg(1, DEFAULT_SAVAGES_COUNT)
	maxServingsCount = getArg(2, DEFAULT_SERVINGS_COUNT)
	M = maxServingsCount
	L = getArg(3, DEFAULT_COOKS_COUNT)
	threadCount = N + L
	mutex = NewMutex(threadCount)

	runtime.GOMAXPROCS(DEFAULT_THREAD_COUNT)
	fmt.Println(CYAN, "\t", "Savages: ", "\t", N)
	fmt.Println(MAGENTA, "\tServings Count:\t", M)
	fmt.Println(GREEN, "\t", "Cooks: ", "\t", L)
	fmt.Println(RESET)

	var i uint
	for i = 0; i < L; i++ {
		go Cook(i)
	}

	for i = 0; i < N; i++ {
		go Savage(i + DEFAULT_COOKS_COUNT)
	}

	// Stops the program after some delay, so the user can see the output
	<- time.After(250 * time.Millisecond)
}

func PrintCook(s string) {
	fmt.Println(GREEN + s + RESET)
}

func PrintSavage(s string) {
	fmt.Println(CYAN + s + RESET)
}

func PrintAlert(s string) {
	fmt.Println(RED + s + RESET)
}

type Mutex struct {
	number []int
}

func NewMutex(size uint) *Mutex {
	return &Mutex{
		number: make([]int, size),
	}
}

func (m *Mutex) maxNumber() int {
	if len(m.number) == 0 {
		return -1
	}

	max := m.number[0]
	for i := 1; i < len(m.number); i++ {
		if m.number[i] > max {
			max = m.number[i]
		}
	}
	return max
}

func (m *Mutex) Lock(i uint) {
	m.number[i] = 1 + m.maxNumber()
	var j uint
	for j = 0; j < threadCount; j++ {
		for m.number[j] != 0 && (m.number[j] < m.number[i] || (m.number[j] == m.number[i] && j < i)) {
			// Wait until all threads with smaller numbers or with the same
			// number, but with higher priority, finish their work
		}
	}
}

func (m *Mutex) Unlock(i uint) {
	m.number[i] = 0
}

type Semaphore struct {
	v    int           // valor do semaforo: negativo significa proc bloqueado
	fila chan struct{} // canal para bloquear os processos se v < 0
	sc   chan struct{} // canal para atomicidade das operacoes wait e signal
}

func NewSemaphore(init int) *Semaphore {
	s := &Semaphore{
		v:    init,                   // valor inicial de creditos
		fila: make(chan struct{}),    // canal sincrono para bloquear processos
		sc:   make(chan struct{}, 1), // usaremos este como semaforo para SC, somente 0 ou 1
	}
	return s
}

func (s *Semaphore) Wait() {
	s.sc <- struct{}{} // SC do semaforo feita com canal
	s.v--              // decrementa valor
	if s.v < 0 {       // se negativo era 0 ou menor, tem que bloquear
		<-s.sc               // antes de bloq, libera acesso
		s.fila <- struct{}{} // bloqueia proc
	} else {
		<-s.sc // libera acesso
	}
}

func (s *Semaphore) Signal() {
	s.sc <- struct{}{} // entra sc
	s.v++
	if s.v <= 0 { // tem processo bloqueado ?
		<-s.fila // desbloqueia
	}
	<-s.sc // libera SC para outra op
}