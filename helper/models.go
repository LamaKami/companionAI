package helper

type ModelTypes struct {
	ModelTypes []ModelType `json:"modelTypes"`
}

type ModelType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type NewModel struct {
	Name string
	Type string
}

type EntityDataPoints struct {
	EntityDataPoints []EntityDataPoint `json:"dataPoints"`
}

type EntityDataPoint struct {
	Id       string              `json:"id"`
	Sentence string              `json:"sentence"`
	Entities []EntityInformation `json:"entities"`
}

type EntityInformation struct {
	StartingPosition int    `json:"start"`
	EndingPosition   int    `json:"end"`
	EntityType       string `json:"type"`
}
