package model

type OrchestrationMetadata struct {
	GofiberMetadata
	Authorization            string
	OrchestrationQueryParams interface{}
	CorrelationID            string
	RequestType              string
}

type YamlWorkload struct {
	StartEvent       interface{}            `yaml:"startEvent"`
	ServiceTask      map[string]interface{} `yaml:"serviceTask"`
	Name             string                 `yaml:"name"`
	Path             string                 `yaml:"path"`
	Method           string                 `yaml:"method"`
	Timeout          int                    `yaml:"timeout"`
	DebugMode        bool                   `yaml:"debugMode"`
	UniqueConstraint *UniqueConstraint      `yaml:"unique_constraint" json:"unique_constraint"`
}

type UniqueConstraint struct {
	Key        string `json:"key"`
	TimeToLive int32  `json:"time_to_live" yaml:"time_to_live"`
}

type YamlWorker struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	ListenTopic string `yaml:"listenTopic"`
	ReplyTopic  string `yaml:"replyTopic"`
	Timeout     int    `yaml:"timeout"`
}
type Service struct {
	Data              interface{}            `json:"data"`
	SourceRef         interface{}            `json:"sourceRef"`
	Condition         interface{}            `json:"condition"`
	SaveGlobal        map[string]interface{} `yaml:"save_global" json:"save_global"`
	DefaultTask       string                 `json:"defaultTask"`
	Worker            string                 `json:"worker"`
	Key               string                 `json:"key"`
	Tipe              string                 `json:"type"`
	EndParsing        string                 `json:"endParsing"`
	RequestType       string                 `json:"request_type" yaml:"request_type"`
	SourceRefParsing  string                 `json:"sourceRefParsing"`
	TargetRef         []interface{}          `json:"targetRef"`
	Retry             int                    `json:"retry"`
	Timeout           int                    `json:"timeout"`
	DefaultStatusCode int                    `json:"defaultStatusCode"`
	Delay             int                    `json:"delay"`
}

type Condition struct {
	VariableCheck []string         `json:"variableCheck"`
	Value         []ConditionValue `json:"value"`
}

type ConditionValue struct {
	SaveGlobal map[string]interface{} `yaml:"save_global" json:"save_global"`
	Target     string                 `json:"target"`
	Case       []interface{}          `json:"case"`
	StatusCode int                    `json:"status_code,omitempty"`
}

type Endpoints struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}
