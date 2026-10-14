package k8s

type ProbeTargetSpec struct {
	URL                 string `json:"url"`
	IntervalSeconds     int    `json:"intervalSeconds"`
	TimeoutSeconds      int    `json:"timeoutSeconds"`
	ExpectedStatusCodes []int  `json:"expectedStatusCodes,omitempty"`
	FailureThreshold    int    `json:"failureThreshold"`
	Active              bool   `json:"active"`
}

type ProbeTarget struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec ProbeTargetSpec `json:"spec"`
}
