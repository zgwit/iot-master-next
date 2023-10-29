package model

type DeviceModelType struct {
	Name string `form:"name" bson:"name" json:"name"`
	Id   string `form:"id" bson:"id" json:"id"`

	Attributes map[string]DeviceModelAttributeType `form:"attributes" bson:"attributes" json:"attributes"`
	Events     map[string]DeviceModelEventType     `form:"events" bson:"events" json:"events"`
	Actions    map[string]DeviceModelActionType    `form:"actions" bson:"actions" json:"actions"`

	KeepAlive int64  `form:"keep_alive" bson:"keep_alive" json:"keep_alive"`
	Script    string `form:"script" bson:"script" json:"script"`

	Drive  string      `form:"drive" bson:"drive" json:"drive"`
	Config interface{} `form:"config" bson:"config" json:"config"`
}

type DeviceModelAttributeType struct {
	Name        string `form:"name" bson:"name" json:"name"`
	DataType    string `form:"data_type" bson:"data_type" json:"data_type"`
	LastestOnly bool   `form:"lastest_only" bson:"lastest_only" json:"lastest_only"`
}

type DeviceModelEventType struct {
	Name string `form:"name" bson:"name" json:"name"`
}

type DeviceModelActionType struct {
	Name string `form:"name" bson:"name" json:"name"`
}
