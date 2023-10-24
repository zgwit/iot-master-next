package model

type KeyValueType map[string]interface{}
type KeyValueStringType map[string]string
type KeyValueNumberType map[string]int64
type KeyValueFloatType map[string]float64
type KeyValueBoolType map[string]bool

type TimeValueType struct {
	Time  int64       `form:"time" bson:"time" json:"time"`
	Value interface{} `form:"value" bson:"value" json:"value"`
}

type TimeValueStringType struct {
	Time  int64  `form:"time" bson:"time" json:"time"`
	Value string `form:"value" bson:"value" json:"value"`
}

type TimeValueBoolType struct {
	Time  int64 `form:"time" bson:"time" json:"time"`
	Value bool  `form:"value" bson:"value" json:"value"`
}

type NameIdType struct {
	Name string `form:"name" bson:"name" json:"name"`
	Id   string `form:"id" bson:"id" json:"id"`
}

type PageType struct {
	Start int64 `form:"start" bson:"start" json:"start"`
	Stop  int64 `form:"stop" bson:"stop" json:"stop"`

	Size   int64 `form:"size" bson:"size" json:"size"`
	Offset int64 `form:"offset" bson:"offset" json:"offset"`

	Desc bool `form:"desc" bson:"desc" json:"desc"`
}
