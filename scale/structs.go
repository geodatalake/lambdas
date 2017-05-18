package scale

import (
	"encoding/json"
)

type GeoMetadata struct {
	Started string          `json:"data_started"`
	Ended   string          `json:"data_ended"`
	GeoJson json.RawMessage `json:"geo_json"`
}

type OutputFile struct {
	Path        string       `json:"path"`
	GeoMetadata *GeoMetadata `json:"geo_metadata,omitempty"`
}

type OutputData struct {
	Name  string        `json:"name"`
	File  *OutputFile   `json:"file,omitempty"`
	Files []*OutputFile `json:"files,omitempty"`
}

type ParseResult struct {
	Filename         string       `json:"filename"`
	NewWorkspacePath string       `json:"new_workspace_path"`
	DataTypes        []string     `json:"data_types"`
	GeoMetadata      *GeoMetadata `json:"geo_json,omitempty"`
}

type ResultsManifest struct {
	Version      string         `json:"version"`
	OutputData   []*OutputData  `json:"output_data,omitempty"`
	ParseResults []*ParseResult `json:"parse_results,omitempty"`
}

type JobEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type JobMount struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Required bool   `json:"required"`
	Mode     string `json:"mode"`
}

// Job Settings are passed as command line arguments
// Their Name MUST match a name in CommandArgs
type JobSetting struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Secret   bool   `json:"secret"`
}

type InputType string

const (
	IProperty InputType = "property"
	IFile     InputType = "file"
	IFiles    InputType = "files"
)

type OutputType string

const (
	OFile  OutputType = "file"
	OFiles OutputType = "files"
)

type JobInput struct {
	Name       string    `json:"name"`
	Type       InputType `json:"type"`
	Required   bool      `json:"required"`
	Partial    bool      `json:"partial"`
	MediaTypes []string  `json:"media_types"`
}

type JobOutput struct {
	Name      string     `json:"name"`
	Type      OutputType `json:"type"`
	Required  bool       `json:"required"`
	MediaType string     `json:"media_type"`
}

// Command Args must match with a Settings property
// or an Input name
// Version 1.4
type JobTypeInterface struct {
	Version         string        `json:"version"`
	Command         string        `json:"command"`
	CommandArgs     string        `json:"command_arguments"`
	EnvVars         []*JobEnvVar  `json:"env_vars,omitempty"`
	Mounts          []*JobMount   `json:"mounts,omitempty"`
	Settings        []*JobSetting `json:"settings,omitempty"`
	InputData       []*JobInput   `json:"input_data"`
	OutputData      []*JobOutput  `json:"output_data"`
	SharedResources []string      `json:"shared_resources"`
}

// Return a JobType only intitalized with Version
// The struct is too complicated for commandline args
func NewRawJobType() *JobTypeInterface {
	return &JobTypeInterface{Version: "1.4"}
}

type BrokerType string

const (
	HostBrokerT BrokerType = "host"
	NfsBrokerT  BrokerType = "nfs"
	S3BrokerT   BrokerType = "s3"
)

type AwsCredentials struct {
	AccessKey string `json:"access_key_id"`
	SecretKey string `json:"secret_access_key"`
}

// HostPath is optional and only used when partial is true
// in the job interface. It is used when s3fs is used to mount
// the bucket path in all nodes. Only read operations are performed
// on the mount.
type S3Broker struct {
	Type        BrokerType      `json:"type"`
	BucketName  string          `json:"bucket_name"`
	Credentials *AwsCredentials `json:"credentials,omitempty"`
	HostPath    string          `json:"host_path,omitempty"`
	Region      string          `json:"region_name,omitempty"`
}

func NewS3Broker(region, bname string) *S3Broker {
	return &S3Broker{Type: S3BrokerT, BucketName: bname, Region: region}
}

// version 1.0
type S3Workspace struct {
	Version string    `json:"version,omitempty"`
	Broker  *S3Broker `json:"broker"`
}

func NewS3Workspace(region, bname string) *S3Workspace {
	return &S3Workspace{Version: "1.0", Broker: NewS3Broker(region, bname)}
}

type CreateWorkspace struct {
	Name        string       `json:"name"`
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	BaseUrl     string       `json:"base_url,omitempty"`
	Active      bool         `json:"is_active,omitempty"`
	Workspace   *S3Workspace `json:"json_config"`
}

type ErrorMapping struct {
	Version   string            `json:"version"`
	ExitCodes map[string]string `json:"exit_codes"`
}

type TriggerCondition struct {
	MediaType string   `json:"media_type"`
	DataTypes []string `json:"data_types,omitempty"`
}

type TriggerConfiguration struct {
	Version   string            `json:"version"`
	Condition *TriggerCondition `json:"condition"`
	Data      map[string]string `json:"data"`
}

type TriggerRule struct {
	Type          string                `json:"type"`
	Active        bool                  `json:"is_active"`
	Configuration *TriggerConfiguration `json:"configuration"`
}

type CreateJob struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Title                string            `json:"title,omitempty"`
	Description          string            `json:"description,omitempty"`
	Category             string            `json:"category,omitempty"`
	AutherName           string            `json:"author_name,omitempty"`
	AutherUrl            string            `json:"auther_url,omitempty"`
	LongRunning          bool              `json:"is_long_running,omitempty"`
	Operational          bool              `json:"is_operational,omitempty"`
	Paused               bool              `json:"is_paused,omitempty"`
	IconCode             string            `json:"icon_code,omitempty"`
	DockerPrivileged     bool              `json:"docker_privileged,omitempty"`
	DockerImage          string            `json:"docker_image,omitempty"`
	Priority             int               `json:"priority,omitempty"`
	Timeout              int               `json:"timeout,omitempty"`
	MaxScheduled         int               `json:"max_scheduled,omitempty"`
	MaxTries             int               `json:"max_tries,omitempty"`
	CpusRequired         float64           `json:"cpus_required,omitempty"`
	MemRequired          float64           `json:"mem_required,omitempty"`
	SharedMemRequired    float64           `json:"shared_mem_required,omitempty"`
	DiskOutConstRequired float64           `json:"disk_out_const_required,omitempty"`
	DiskOutMultRequired  float64           `json:"disk_out_mult_required,omitempty"`
	Interface            *JobTypeInterface `json:"interface"`
	Errors               *ErrorMapping     `json:"error_mapping,omitempty"`
	Trigger              *TriggerRule      `json:"trigger_rule,omitempty"`
}

type BoundsResult struct {
	Bounds string `json:"bounds"`
	Prj    string `json:"prj"`
}
