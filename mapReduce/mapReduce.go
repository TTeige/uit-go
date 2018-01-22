package mapReduce

type MapCollector chan chan interface{}

//Function that is executed during the mapping step
type MapperFunc func(interface{}, chan interface{})

//Function that is executed during the reduce step
type ReducerFunc func(chan interface{}, chan interface{})

type FilterFunc func(interface{}) bool

type IMapReducer interface {
	New(int) *MapReducer
	Filter(FilterFunc)
	Map(MapperFunc)
	Reduce(ReducerFunc)
	Run() interface{}
}

type MapReducer struct {
	mapFunc     MapperFunc
	reducerFunc ReducerFunc
	filterFunc  FilterFunc
	collector   MapCollector
	step        int
	maxWorkers  int
}

func New(maxWorkers int) *MapReducer {
	return &MapReducer{
		step:       0,
		maxWorkers: maxWorkers,
	}
}

func (m *MapReducer) Filter(filterFunc FilterFunc) {
	m.filterFunc = filterFunc
}

func (m *MapReducer) Map(mapperFunc MapperFunc) {
	m.mapFunc = mapperFunc
}

func (m *MapReducer) Reduce(reducerFunc ReducerFunc) {
	m.reducerFunc = reducerFunc
}

func (m *MapReducer) Run(input chan interface{}) interface{} {
	collector := make(MapCollector, m.maxWorkers)
	reducerInput := make(chan interface{})
	reducerOutput := make(chan interface{})

	go m.reducerFunc(reducerInput, reducerOutput)
	go m.reducerDispatcher(reducerInput, collector)
	go m.mapperDispatcher(input, collector, m.mapFunc)

	return <-reducerOutput

}
func (m *MapReducer) mapperDispatcher(input chan interface{}, collector MapCollector, mapperFunc MapperFunc) {
	if mapperFunc == nil {
		collector <- input
		close(collector)
		return
	}
	for item := range input {
		mapperOutput := make(chan interface{})
		go mapperFunc(item, mapperOutput)
		collector <- mapperOutput
	}
	close(collector)
}

func (m *MapReducer) reducerDispatcher(reducerInput chan interface{}, collector MapCollector) {
	for mapperOutput := range collector {
		reducerInput <- <-mapperOutput
	}
	close(reducerInput)
}
