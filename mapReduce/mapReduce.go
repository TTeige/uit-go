package mapReduce

import (
	"strings"
	"io"
	"bufio"
	"os"
	"log"
)

type MapCollector chan chan interface{}

//Function that is executed during the mapping step
type MapperFunc func(interface{}, chan interface{})

//Function that is executed during the reduce step
type ReducerFunc func(chan interface{}, chan interface{})

type MapReducer struct {
	mapFunc      MapperFunc
	mapCollector MapCollector
	reducerFunc  ReducerFunc
	collector    MapCollector
}

func NewMapReducer(mapperFunc MapperFunc, reducerFunc ReducerFunc, maxWorkers int) *MapReducer {
	return &MapReducer{
		mapFunc:     mapperFunc,
		reducerFunc: reducerFunc,
		collector:   make(MapCollector, maxWorkers),
	}
}

//Input is a channel of input values that are parsed. These values are used in the mapping step of the mapreduce algorithm.
func (m *MapReducer) MapReduce(input chan interface{}) interface{} {
	reducerInput := make(chan interface{})
	reducerOutput := make(chan interface{})

	go m.reducerFunc(reducerInput, reducerOutput)
	go m.reducerDispatcher(reducerInput)
	go m.mapperDispatcher(input)

	return <-reducerOutput
}

func (m *MapReducer) mapperDispatcher(input chan interface{}) {
	for item := range input {
		mapperOutput := make(chan interface{})
		go m.mapFunc(item, mapperOutput)
		m.collector <- mapperOutput
	}
	close(m.collector)
}

func (m *MapReducer) reducerDispatcher(reducerInput chan interface{}) {
	for output := range m.collector {
		reducerInput <- <-output
	}
	close(reducerInput)
}

func EnumerateJSON(filename string) chan string {
	output := make(chan string)
	go func() {
		log.Println("?")
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				log.Printf("%s\n", err)
				break
			} else if err != nil {
				log.Fatalf("%s\n", err)
			}
			if strings.HasPrefix(line, "#") == true {
				continue
			}
			log.Printf("Line: %s\n", line)
			output <- line
		}
		close(output)
	}()
	return output
}
