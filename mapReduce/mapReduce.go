package mapReduce

type MapCollector chan chan interface{}

//Function that is executed during the mapping step
type MapperFunc func(interface{}, chan interface{})

//Function that is executed during the reduce step
type ReducerFunc func(chan interface{}, chan interface{})

//
type IMapReducer interface {
	New(int) *MapReducer
	Filter(interface{}) bool
	Map(MapperFunc)
	Reduce(ReducerFunc)
	Run() interface{}

	NewMapReducer(MapperFunc, ReducerFunc, int) *MapReducer
	MapReduce(chan interface{}) interface{}
}

type MapReducer struct {
	mapFunc     MapperFunc
	reducerFunc ReducerFunc
	collector   MapCollector
}

//Constructs a new map reducer uses the given functions
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
