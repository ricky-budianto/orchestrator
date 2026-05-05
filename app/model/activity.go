package model

type ActivityDetails struct {
	EndpointPath   string                 `json:"endpoint_path,omitempty"`               // if protocol is HTTP
	HTTPMethod     string                 `json:"http_method" gorm:"column:http_method"` // if protocol is HTTP
	RequestType    string                 `json:"request_type,omitempty"`                // if protocol is AMQP/Kafka
	RequestData    interface{}            `json:"request_data,omitempty"`                // request data
	ResponseData   interface{}            `json:"response_data,omitempty"`               // response data
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`             // additional info
}

type Activity struct {
	ID                    string          `json:"id" gorm:"type:varchar(36);default:uuid_generate_v4()"`
	Name                  string          `json:"name" gorm:"type:varchar(255);index"`             // name of the activity
	CommunicationProtocol string          `json:"communication_protocol" gorm:"type:varchar(100)"` // HTTP or AMQP or GRPC or TCP
	SourceIPAddress       string          `json:"source_ip_address" gorm:"type:varchar(100)"`      // source IP address
	ActivityDetails       ActivityDetails `json:"activity_details" gorm:"serializer:json"`
}
