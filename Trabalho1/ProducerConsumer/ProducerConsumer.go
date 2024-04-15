package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

const (
	PRODUCERS_COUNT = 2
	CONSUMERS_COUNT = 6
	THREAD_COUNT = PRODUCERS_COUNT + CONSUMERS_COUNT
	QUEUE_SIZE = 3

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
	mutex *Mutex = NewMutex()
	queue  Queue = make(Queue, 0, QUEUE_SIZE)
)

// Returns a random integer between 0 and max
func GetRandom(max int) int {
	return rand.Intn(max + 1)
}

// Waits for a random time in milliseconds with 10x a custom multiplier
func WaitRandom(multiplier int) {
	time.Sleep(time.Duration(GetRandom(10 * multiplier)) * time.Millisecond)
}

// @deprecated
// WORKAROUND
// This is a way of showing the user the aproximate time the operation was done contrary to the time the operation was printed
// This is a workaround because, sometimes, a thread has already inserted/removed an item from the queue but the OS gives the CPU to another thread to remove another item, so printing goes out of order
func CurrentTimeStamp() string {
	milliseconds := time.Now().UnixMilli()
	milliseconds = milliseconds % 10_000_000
	return fmt.Sprint(milliseconds)
}

func Producer(id int) {
	for {
		mutex.Lock(id)
			fmt.Println(YELLOW, id, "- Pronto para produzir...", RESET)
			item := GetRandom(99)
			success := queue.Enqueue(item)
			if success {
				fmt.Println(GREEN,
					"\tAdicionando: ", item, "\n",
					"\tFila Seção Crítica: ", queue, "\n",
					"\tTimeStamp: ", CurrentTimeStamp(), "\n",
				RESET)
			} else {
				fmt.Println(RED, "\tFila cheia, não foi possível adicionar\n", RESET)
			}
		mutex.Unlock(id)
		WaitRandom(PRODUCERS_COUNT)
	}
}

func Consumer(id int) {
	for {
		WaitRandom(CONSUMERS_COUNT)
		mutex.Lock(id)
			fmt.Println(YELLOW, id, "- Tentando consumir...", RESET)
			previousQueue := queue
			item, success := queue.Dequeue()
			if success {
				fmt.Println(CYAN,
					"\tAnterior: ", previousQueue, "\n",
					"\tRemovido: ", item, "\n",
					"\tNew: ", queue, "\n",
					"\tTimeStamp: ", CurrentTimeStamp(), "\n",
				RESET)
			} else {
				fmt.Println(RED, "\tFila vazia, não foi possível remover\n", RESET)
			}
		mutex.Unlock(id)
	}
}

func main() {
	runtime.GOMAXPROCS(THREAD_COUNT)
	fmt.Println(RESET)
	fmt.Println(BOLD, "Producer Consumer with: ")
	fmt.Println(MAGENTA, "\tQueue Size:\t", QUEUE_SIZE)
	fmt.Println(GREEN, "\t", "Producers: ", "\t", PRODUCERS_COUNT)
	fmt.Println(CYAN, "\t", "Consumers: ", "\t", CONSUMERS_COUNT)
	fmt.Println(RESET)

	for i := 0; i < PRODUCERS_COUNT; i++ {
		go Producer(i)
	}

	for i := 0; i < CONSUMERS_COUNT; i++ {
		go Consumer(i + PRODUCERS_COUNT)
	}

	// Stops the program after some delay, so the user can see the output
	<- time.After(250 * time.Millisecond)
}

type Queue []int

func (q *Queue) Enqueue(item int) (success bool) {
	if len(*q) == QUEUE_SIZE {
		return false
	}
	*q = append(*q, item)
	return true
}

func (q *Queue) Dequeue() (head int, success bool) {
	if len(*q) == 0 {
		return -1, false
	}
	item := (*q)[0]
	*q = (*q)[1:]
	return item, true
}

type Mutex struct {
	number [THREAD_COUNT]int
}

func NewMutex() *Mutex {
	return &Mutex{
		number: [THREAD_COUNT]int{},
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

func (m *Mutex) Lock(i int) {
	m.number[i] = 1 + m.maxNumber()
	for j := 0; j < THREAD_COUNT; j++ {
		for m.number[j] != 0 && 
		(m.number[j] < m.number[i] || (m.number[j] == m.number[i] && j < i)) {
			// Wait until all threads with smaller numbers or with the same
			// number, but with higher priority, finish their work
		}
	}
}

func (m *Mutex) Unlock(i int) {
	m.number[i] = 0
}