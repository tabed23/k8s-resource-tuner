package models

import (
    "time"
    v1 "k8s.io/api/core/v1"
)

type WorkLoad struct {
    Namespace  string            `json:"namespace"`
    Name       string            `json:"name"`
    Kind       string            `json:"kind"`
    Containers []ContainerSpec   `json:"containers"`
    Labels     map[string]string `json:"labels"`
}

type ContainerSpec struct {
    Name      string         `json:"name"`
    Resources ResourceConfig `json:"resources"`
}

type ResourceConfig struct {
    Request v1.ResourceList `json:"requests,omitempty"`
    Limits  v1.ResourceList `json:"limits,omitempty"`
}

type Usage struct {
    Timestamp time.Time `json:"timestamp"`
    CPU       float64   `json:"cpu"`
    Memory    float64   `json:"memory"`
}

type UsageStats struct {
    ContainerName string    `json:"container_name"`
    CPUSamples    []float64 `json:"cpu_samples"`
    MemSamples    []float64 `json:"mem_samples"`
    CPUAvg        float64   `json:"cpu_avg"`
    CPUP95        float64   `json:"cpu_p95"`
    CPUP99        float64   `json:"cpu_p99"`
    MemAvg        float64   `json:"mem_avg"`
    MemP95        float64   `json:"mem_p95"`
    MemP99        float64   `json:"mem_p99"`
}

type Recommendation struct {
    ContainerName      string         `json:"container_name"`
    RecommendedRequest ResourceConfig `json:"recommended_request"`
    RecommendedLimit   ResourceConfig `json:"recommended_limit"`
    Reason             string         `json:"reason"`
}

type Report struct {
    Timestamp time.Time     `json:"timestamp"`
    Entries   []ReportEntry `json:"entries"`
    Summary   string        `json:"summary"`
}

type ReportEntry struct {
    Workload       WorkLoad         `json:"workload"`
    Stats          []UsageStats     `json:"stats"`
    Recommendation []Recommendation `json:"recommendations"`
}

type SlackMessage struct {
    Channel string `json:"channel"`
    Text    string `json:"text"`
}