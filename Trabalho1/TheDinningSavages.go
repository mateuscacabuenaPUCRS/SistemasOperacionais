package main

import (
	"fmt"
	"time"
)

const (
	Red = "\033[31m"
	Green = "\033[32m"
	Yellow = "\033[33m"
	Reset = "\033[0m"
	/* Number of servings in the pot */
	M = 5
)

var (
	servings = M
	mutex = NewSemaphore(1)
	emptyPot = NewSemaphore(0)
	fullPot = NewSemaphore(0)
)

func Cook() {
	for {
		emptyPot.Wait()
		putServingsInPot()
		fullPot.Signal()
	}
}

func putServingsInPot() {
	PrintCook("Putting servings in pot...")
}

func Savage(id string) {
	for {
		mutex.Wait()
		if servings == 0 {
			wakeUpCook(id)
			emptyPot.Signal()
			fullPot.Wait()
			servings = M
		}
		servings--
		getServingFromPot(id)
		mutex.Signal()
		eat(id)
	}
}

func wakeUpCook(id string) {
	PrintSavage(fmt.Sprintf("%s - Pot is empty, I'll wake up the cook...", id), true)
}

func getServingFromPot(id string) {
	PrintSavage(fmt.Sprintf("%s is serving...", id), false)
}

func eat(id string) {
	PrintSavage(fmt.Sprintf("%s is eating...", id), false)
}

func main() {
	fmt.Println("Dining Savages")
	for i := 0; i < 1; i++ {
		fmt.Println("Making cook ", i)
		go Cook()
	}
	for i := 0; i < 5; i++ {
		fmt.Println("Making savage ", i)
		go Savage(fmt.Sprintf("%d", i))
	}
	<- time.After(50 * time.Millisecond)
}

func PrintCook(s string) {
	fmt.Println(Red + s + Reset)
}

func PrintSavage(s string, alert bool) {
	if alert {
		fmt.Println(Yellow + s + Reset)
	} else {
		fmt.Println(Green + s + Reset)
	}
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