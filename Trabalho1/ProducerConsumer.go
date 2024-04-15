package main

import (
    "errors"
    "fmt"
)

const tamanhoMaximo = 5

type Vetor struct {
    elementos []int
}

func (v *Vetor) AdicionarElemento(elemento int) error {
    if len(v.elementos) >= tamanhoMaximo {
        return errors.New("Vetor cheio, não é possível adicionar mais elementos")
    }

    v.elementos = append(v.elementos, elemento)
    return nil
}

func (v *Vetor) RemoverElemento() (int, error) {
	if len(v.elementos) == 0 {
		return 0, errors.New("Vetor vazio, não é possível remover elementos")
	}

	elementoRemovido := v.elementos[0]
	v.elementos = v.elementos[1:]
	return elementoRemovido, nil
}

func (v *Vetor) next() (int, error) {
	if len(v.elementos) == 0 {
		return 0, errors.New("Vetor vazio, não há próximo elemento na fila")
	}

	return v.elementos[0], nil
}

func main() {
    vetor := Vetor{}

    for i := 0; i < 6; i++ {
        if err := vetor.AdicionarElemento(i + 1); err != nil {
            fmt.Println(err)
        } else {
            fmt.Println("Elemento adicionado ao vetor")
			fmt.Println(vetor.elementos)
        }
    }

	// vetor.RemoverElemento()
	// fmt.Println("Elemento removido do vetor")
	// fmt.Println(vetor.elementos)

	fmt.Println(vetor.next())
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
