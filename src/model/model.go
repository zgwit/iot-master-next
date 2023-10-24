package zgwit_model

type KeyValue map[string]interface{}
type KeyValueString map[string]string
type KeyValueNumber map[string]int64
type KeyValueFloat map[string]float64
type KeyValueBool map[string]bool

type TimeValue struct {
	Time  int64       `form:"time" bson:"time" json:"time"`
	Value interface{} `form:"value" bson:"value" json:"value"`
}

type TimeValueString struct {
	Time  int64  `form:"time" bson:"time" json:"time"`
	Value string `form:"value" bson:"value" json:"value"`
}

type TimeValueBool struct {
	Time  int64 `form:"time" bson:"time" json:"time"`
	Value bool  `form:"value" bson:"value" json:"value"`
}

type NameId struct {
	Name string `form:"name" bson:"name" json:"name"`
	Id   string `form:"id" bson:"id" json:"id"`
}

type Page struct {
	Start int64 `form:"start" bson:"start" json:"start"`
	Stop  int64 `form:"stop" bson:"stop" json:"stop"`

	Size   int64 `form:"size" bson:"size" json:"size"`
	Offset int64 `form:"offset" bson:"offset" json:"offset"`

	Desc bool `form:"desc" bson:"desc" json:"desc"`
}
