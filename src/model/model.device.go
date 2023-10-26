package model

type DeviceType struct {
	Name    string `form:"name" bson:"name" json:"name"`
	Id      string `form:"id" bson:"id" json:"id"`
	ModelId string `form:"model_id" bson:"model_id" json:"model_id"`

	Labels map[string]string `form:"labels" bson:"labels" json:"labels"`

	Drive       string      `form:"drive" bson:"drive" json:"drive"`
	DriveConfig interface{} `form:"drive_config" bson:"drive_config" json:"drive_config"`
}
